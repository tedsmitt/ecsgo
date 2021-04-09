package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/viper"
)

// ExecCommand is a struct that contains information about our command state
type ExecCommand struct {
	cmd        chan string
	err        chan error
	done       chan bool
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

// CreateExecCommand initialises a new ExecCommand struct with the required
// initial values
func CreateExecCommand() *ExecCommand {
	client := createEcsClient()
	e := &ExecCommand{
		cmd:      make(chan string, 1),
		err:      make(chan error, 1),
		done:     make(chan bool),
		client:   client,
		region:   client.SigningRegion,
		endpoint: client.Endpoint,
	}
	return e
}

// Start begins a goroutine that listens on the cmd channel for instructions
func (e *ExecCommand) Start() {
	go func() {
		for {
			select {
			case cmd := <-e.cmd:
				switch cmd {
				case "getCluster":
					e.getCluster()
				case "getService":
					e.getService()
				case "getTask":
					e.getTask()
				case "getContainer":
					e.getContainer()
				case "executeCommand":
					e.executeCmd()
				default:
					e.getCluster()
				}
			case err := <-e.err:
				fmt.Printf(red("\n%s\n"), err)
				os.Exit(1)
			}
		}
	}()
	// Initialise the workflow
	e.cmd <- "getCluster"
	// Block until we receive a message on the done channel
	<-e.done
}

// Lists available clusters and prompts the user to select one
func (e *ExecCommand) getCluster() {
	list, err := e.client.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		e.err <- err
		return
	}
	if len(list.ClusterArns) > 0 {
		var clusterNames []string
		for _, c := range list.ClusterArns {
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
		e.cmd <- "getService"
		return
	} else {
		err := errors.New("No clusters found in account or region")
		e.err <- err
		return
	}
}

// Lists available service and prompts the user to select one
func (e *ExecCommand) getService() {
	list, err := e.client.ListServices(&ecs.ListServicesInput{
		Cluster: aws.String(e.cluster),
	})
	if err != nil {
		e.err <- err
		return
	}
	if len(list.ServiceArns) > 0 {
		var serviceNames []string
		for _, c := range list.ServiceArns {
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
			e.cmd <- "getCluster"
			return
		}
		e.service = selection
		e.cmd <- "getTask"
		return
	} else {
		e.err <- err
		return
	}
}

// Lists tasks in a cluster and prompts the user to select one
func (e *ExecCommand) getTask() {
	var input *ecs.ListTasksInput
	if e.service == "*" {
		input = &ecs.ListTasksInput{
			Cluster: aws.String(e.cluster),
		}
	} else {
		input = &ecs.ListTasksInput{
			Cluster:     aws.String(e.cluster),
			ServiceName: aws.String(e.service),
		}
	}
	list, err := e.client.ListTasks(input)
	if err != nil {
		e.err <- err
		return
	}
	e.tasks = make(map[string]*ecs.Task)
	if len(list.TaskArns) > 0 {
		describe, err := e.client.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: aws.String(e.cluster),
			Tasks:   list.TaskArns,
		})
		if err != nil {
			e.err <- err
			return
		}
		// Add each task to a map for easier retrieval
		for _, t := range describe.Tasks {
			taskId := strings.Split(*t.TaskArn, "/")[2]
			e.tasks[taskId] = t
		}
		// Ask the user to select which Task to connect to
		selection, err := selectTask(e.tasks)
		if err != nil {
			e.err <- err
			return
		}
		if *selection.TaskArn == backOpt {
			e.cmd <- "getService"
			return
		}
		e.task = selection
		e.cmd <- "getContainer"
		return
	} else {
		err := errors.New(fmt.Sprintf("There are no running tasks in the cluster %s", e.cluster))
		e.err <- err
		return
	}
}

// Lists containers in a task and prompts the user to select one (if there is more than 1 container)
// otherwise returns the the only container in the task
func (e *ExecCommand) getContainer() {
	if len(e.task.Containers) > 1 {
		// Ask the user to select a container
		selection, err := selectContainer(e.task.Containers)
		if err != nil {
			e.err <- err
			return
		}
		if *selection.Name == backOpt {
			e.cmd <- "getTask"
			return
		}
		e.container = selection
		e.cmd <- "executeCommand"
		return
	} else {
		// There is only one container in the task, return it
		e.container = e.task.Containers[0]
		e.cmd <- "executeCommand"
		return
	}
}

// executeCmd takes all of our previous values and builds a session for us
// and then calls runCommand to execute the session input via session-manager-plugin
func (e *ExecCommand) executeCmd() {
	// Check if command has been passed to the tool, otherwise default
	// to /bin/sh
	var command string
	if viper.GetString("cmd") != "" {
		command = viper.GetString("cmd")
	} else {
		command = "/bin/sh"
	}

	execCommand, err := e.client.ExecuteCommand(&ecs.ExecuteCommandInput{
		Cluster:     aws.String(e.cluster),
		Interactive: aws.Bool(true),
		Task:        e.task.TaskArn,
		Command:     aws.String(command),
		Container:   e.container.Name,
	})
	if err != nil {
		e.err <- err
		return
	}
	execSess, err := json.Marshal(execCommand.Session)
	if err != nil {
		e.err <- err
		return
	}
	target := ssm.StartSessionInput{
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", e.cluster, e.task, *e.container.RuntimeId)),
	}
	targetJson, err := json.Marshal(target)
	if err != nil {
		e.err <- err
		return
	}

	fmt.Printf("\nCluster: %v | Service: %v | Task: %s", cyan(e.cluster), magenta(e.service), green(strings.Split(*e.task.TaskArn, "/")[2]))
	fmt.Printf("\nConnecting to container %v", yellow(*e.container.Name))
	// Expecting session-manager-plugin to be found in $PATH
	if err = runCommand("session-manager-plugin", string(execSess), e.region, "StartSession", "", string(targetJson), e.endpoint); err != nil {
		e.done <- true
		e.err <- err
		return
	}

	/* 	if connect := promptConnectAgain(); connect == true {
		e.cmd <- "getTask"
		return
	} */

	e.done <- true
	return
}
