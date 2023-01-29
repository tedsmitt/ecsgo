package app

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/spf13/viper"
)

// App is a struct that contains information about our command state
type App struct {
	input      chan string
	err        chan error
	exit       chan error
	client     ecsiface.ECSAPI
	region     string
	endpoint   string
	cluster    string
	service    string
	task       *ecs.Task
	tasks      map[string]*ecs.Task
	container  *ecs.Container
	containers []*ecs.Container
}

// CreateApp initialises a new App struct with the required initial values
func CreateApp() *App {
	client := createEcsClient()
	e := &App{
		input:    make(chan string, 1),
		err:      make(chan error, 1),
		exit:     make(chan error, 1),
		client:   client,
		region:   client.SigningRegion,
		endpoint: client.Endpoint,
	}

	return e
}

// Start begins a goroutine that listens on the input channel for instructions
func (e *App) Start() error {
	// Before we do anything make sure that the session-manager-plugin is available in $PATH, exit if it isn't
	_, err := exec.LookPath("session-manager-plugin")
	if err != nil {
		fmt.Println(Red("session-manager-plugin isn't installed or wasn't found in $PATH - https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html"))
		return err
	}

	go func() {
		for {
			select {
			case input := <-e.input:
				switch input {
				case "getCluster":
					e.getCluster()
				case "getService":
					e.getService()
				case "getTask":
					e.getTask()
				case "getContainer":
					e.getContainer()
				case "execute":
					if viper.GetBool("forward") {
						e.executeForward()
					} else {
						e.executeCommand()
					}
				default:
					e.getCluster()
				}
			case err := <-e.err:
				e.exit <- err
			}
		}
	}()

	// Initiate the workflow
	e.input <- "getCluster"

	// Block until we receive a message on the exit channel
	err = <-e.exit
	if err != nil {
		return err
	}

	return nil
}

// Lists available clusters and prompts the user to select one
func (e *App) getCluster() {
	var clusters []*string
	var nextToken *string

	cliArg := viper.GetString("cluster")
	if cliArg != "" {
		e.cluster = cliArg
		e.input <- "getService"
		viper.Set("cluster", "") // Reset the cli arg so user can navigate
		return
	}

	list, err := e.client.ListClusters(&ecs.ListClustersInput{
		MaxResults: awsMaxResults,
	})
	if err != nil {
		e.err <- err
		return
	}
	clusters = append(clusters, list.ClusterArns...)
	nextToken = list.NextToken

	if nextToken != nil {
		for {
			list, err := e.client.ListClusters(&ecs.ListClustersInput{
				MaxResults: awsMaxResults,
				NextToken:  nextToken,
			})
			if err != nil {
				e.err <- err
				return
			}
			clusters = append(clusters, list.ClusterArns...)
			if list.NextToken == nil {
				break
			} else {
				nextToken = list.NextToken
			}
		}
	}

	// Sort the list of clusters alphabetically
	sort.Slice(clusters, func(i, j int) bool {
		return *clusters[i] < *clusters[j]
	})

	if len(clusters) > 0 {
		var clusterNames []string
		for _, c := range clusters {
			arnSplit := strings.Split(*c, "/")
			name := arnSplit[len(arnSplit)-1]
			clusterNames = append(clusterNames, name)
		}

		selection, err := selectCluster(clusterNames)
		if err != nil {
			e.err <- err
			return
		}

		e.cluster = selection
		e.input <- "getService"
		return

	} else {
		err := errors.New("No clusters found in account or region")
		e.err <- err
		return
	}
}

// Lists available services and prompts the user to select one
func (e *App) getService() {
	var services []*string
	var nextToken *string

	cliArg := viper.GetString("service")
	if cliArg != "" {
		e.service = cliArg
		e.input <- "getTask"
		viper.Set("service", "") // Reset the cli arg so user can navigate
		return
	}

	list, err := e.client.ListServices(&ecs.ListServicesInput{
		Cluster:    aws.String(e.cluster),
		MaxResults: awsMaxResults,
	})
	if err != nil {
		e.err <- err
		return
	}
	services = append(services, list.ServiceArns...)
	nextToken = list.NextToken

	if nextToken != nil {
		for {
			list, err := e.client.ListServices(&ecs.ListServicesInput{
				Cluster:    aws.String(e.cluster),
				MaxResults: awsMaxResults,
				NextToken:  nextToken,
			})
			if err != nil {
				e.err <- err
				return
			}
			services = append(services, list.ServiceArns...)
			if list.NextToken == nil {
				break
			} else {
				nextToken = list.NextToken
			}
		}
	}

	// Sort the list of services alphabetically
	sort.Slice(services, func(i, j int) bool {
		return *services[i] < *services[j]
	})

	if len(services) > 0 {
		var serviceNames []string

		for _, c := range services {
			arnSplit := strings.Split(*c, "/")
			name := arnSplit[len(arnSplit)-1]
			serviceNames = append(serviceNames, name)
		}

		selection, err := selectService(serviceNames)
		if err != nil {
			e.err <- err
			return
		}

		if selection == backOpt {
			e.service = ""
			e.input <- "getCluster"
			return
		}

		e.service = selection
		e.input <- "getTask"
		return

	} else {
		// Continue without setting a service if no services are found in the cluster
		fmt.Printf(Yellow("\n%s"), "No services found in the cluster, returning all running tasks...\n")
		e.input <- "getTask"
		return
	}
}

// Lists tasks in a cluster and prompts the user to select one
func (e *App) getTask() {
	var taskArns []*string
	var nextToken *string

	var input *ecs.ListTasksInput

	cliArg := viper.GetString("task")
	if cliArg != "" {
		describe, err := e.client.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: aws.String(e.cluster),
			Tasks:   []*string{aws.String(cliArg)},
		})
		if err != nil {
			e.err <- err
			return
		}
		if len(describe.Tasks) > 0 {
			e.task = describe.Tasks[0]
			e.getContainerOS()
			e.input <- "getContainer"
			viper.Set("task", "") // Reset the cli arg so user can navigate
			return
		} else {
			fmt.Printf(Red(fmt.Sprintf("\nTask with ID %s not found in cluster %s\n", cliArg, e.cluster)))
			e.input <- "getService"
			return
		}
	}

	// If no service has been set, or if ALL (*) services have been selected
	// then we don't need to specify a ServiceName
	if e.service == "" || e.service == "*" {
		input = &ecs.ListTasksInput{
			Cluster:    aws.String(e.cluster),
			MaxResults: awsMaxResults,
		}
	} else {
		input = &ecs.ListTasksInput{
			Cluster:     aws.String(e.cluster),
			ServiceName: aws.String(e.service),
			MaxResults:  awsMaxResults,
		}
	}

	list, err := e.client.ListTasks(input)
	if err != nil {
		e.err <- err
		return
	}

	taskArns = append(taskArns, list.TaskArns...)
	nextToken = list.NextToken

	if nextToken != nil {
		for {
			list, err := e.client.ListTasks(&ecs.ListTasksInput{
				Cluster:    aws.String(e.cluster),
				MaxResults: awsMaxResults,
				NextToken:  nextToken,
			})
			if err != nil {
				e.err <- err
				return
			}
			taskArns = append(taskArns, list.TaskArns...)
			if list.NextToken == nil {
				break
			} else {
				nextToken = list.NextToken
			}
		}
	}

	e.tasks = make(map[string]*ecs.Task)
	if len(taskArns) > 0 {
		describe, err := e.client.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: aws.String(e.cluster),
			Tasks:   taskArns,
		})
		if err != nil {
			e.err <- err
			return
		}

		for _, t := range describe.Tasks {
			taskId := strings.Split(*t.TaskArn, "/")[2]
			e.tasks[taskId] = t
		}

		selection, err := selectTask(e.tasks)
		if err != nil {
			e.err <- err
			return
		}

		if *selection.TaskArn == backOpt {
			e.task = nil
			if e.service == "" {
				e.input <- "getCluster"
				return
			}
			e.input <- "getService"
			return
		}
		e.task = selection
		e.getContainerOS()
		e.input <- "getContainer"
		return

	} else {
		if e.service == "" {
			err := errors.New(fmt.Sprintf("There are no running tasks in the cluster %s\n", e.cluster))
			e.err <- err
			return
		} else {
			fmt.Printf(Red(fmt.Sprintf("\nThere are no running tasks for the service %s in cluster %s\n", e.service, e.cluster)))
			e.input <- "getService"
			return
		}
	}
}

// Lists containers in a task and prompts the user to select one (if there is more than 1 container)
// otherwise returns the the only container in the task
func (e *App) getContainer() {
	cliArg := viper.GetString("container")
	if cliArg != "" {
		for _, c := range e.task.Containers {
			if *c.Name == cliArg {
				e.container = c
				e.input <- "execute"
				return
			}
		}
		fmt.Printf(Red(fmt.Sprintf("\nSupplied container with name %s not found in task %s, cluster %s\n", cliArg, *e.task.TaskArn, e.cluster)))
	}

	if len(e.task.Containers) > 1 {
		selection, err := selectContainer(e.task.Containers)
		if err != nil {
			e.err <- err
			return
		}

		if *selection.Name == backOpt {
			e.input <- "getTask"
			return
		}

		e.container = selection
		e.input <- "execute"
		return

	} else {
		// There is only one container in the task, return it
		e.container = e.task.Containers[0]
		e.input <- "execute"
		return
	}
}

// Determines the OS family of the container instance the task is running on
func (e *App) getContainerOS() {
	// Get associated task definition and determine OS family if EC2 launch-type
	if *e.task.LaunchType == "EC2" {
		family, err := getPlatformFamily(e.client, e.cluster, e.task)
		if err != nil {
			e.err <- err
			return
		}
		// if the OperatingSystemFamily has not been specified in the task definition
		// then we refer to the container instance to determine the OS
		if family == "" {
			ec2Client := createEc2Client()
			family, err = getContainerInstanceOS(e.client, ec2Client, e.cluster, *e.task.ContainerInstanceArn)
			if err != nil {
				e.err <- err
				return
			}
		}
		// Add our own PlatformFamily value for the task struct
		e.task.PlatformFamily = &family
	}
}
