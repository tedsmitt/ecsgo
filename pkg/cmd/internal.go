package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var (
	region   string
	endpoint string

	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

func createEcsClient() *ecs.ECS {
	region := viper.GetString("region")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		Profile:           viper.GetString("profile"),
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := ecs.New(sess)
	region = client.SigningRegion
	endpoint = client.Endpoint

	return client
}

// Lists available clusters and prompts the user to select one
func getCluster(client ecsiface.ECSAPI) (string, error) {
	list, err := client.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		return "", err
	}
	var clusterName string
	if len(list.ClusterArns) > 0 {
		var clusterNames []string
		for _, c := range list.ClusterArns {
			arnSplit := strings.Split(*c, "/")
			name := arnSplit[len(arnSplit)-1]
			clusterNames = append(clusterNames, name)
		}
		selection, err := selectCluster(clusterNames)
		if err != nil {
			return "", err
		}
		clusterName = selection
		return clusterName, nil
	} else {
		err := errors.New("No clusters found in account or region")
		return "", err
	}
}

// Lists available service and prompts the user to select one
func getService(client ecsiface.ECSAPI, clusterName string) (string, error) {
	list, err := client.ListServices(&ecs.ListServicesInput{
		Cluster: aws.String(clusterName),
	})
	if err != nil {
		return "", err
	}
	var serviceName string
	if len(list.ServiceArns) > 0 {
		var serviceNames []string
		for _, c := range list.ServiceArns {
			arnSplit := strings.Split(*c, "/")
			name := arnSplit[len(arnSplit)-1]
			serviceNames = append(serviceNames, name)
		}
		selection, err := selectService(serviceNames)
		if err != nil {
			return "", err
		}
		serviceName = selection
		return serviceName, nil
	} else {
		return "", err
	}
}

// Lists tasks in a cluster and prompts the user to select one
func getTask(client ecsiface.ECSAPI, clusterName string, serviceName string) (*ecs.Task, error) {
	var input *ecs.ListTasksInput
	if serviceName == "*" {
		input = &ecs.ListTasksInput{
			Cluster: aws.String(clusterName),
		}
	} else {
		input = &ecs.ListTasksInput{
			Cluster:     aws.String(clusterName),
			ServiceName: aws.String(serviceName),
		}
	}
	list, err := client.ListTasks(input)
	if err != nil {
		return &ecs.Task{}, err
	}
	if len(list.TaskArns) > 0 {
		describe, err := client.DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: aws.String(clusterName),
			Tasks:   list.TaskArns,
		})
		if err != nil {
			return &ecs.Task{}, err
		}
		// Ask the user to select which Task to connect to
		selection, err := selectTask(describe.Tasks)
		if err != nil {
			return &ecs.Task{}, err
		}
		task := selection
		return task, nil
	} else {
		err := errors.New(fmt.Sprintf("There are no running tasks in the cluster %s", clusterName))
		return &ecs.Task{}, err
	}
}

// Lists containers in a task and prompts the user to select one (if there is more than 1 container)
// otherwise returns the the only container in the task
func getContainer(task *ecs.Task) (*ecs.Container, error) {
	if len(task.Containers) > 1 {
		// Ask the user to select a container
		selection, err := selectContainer(task.Containers)
		if err != nil {
			return &ecs.Container{}, err
		}
		return selection, nil
	} else {
		// There is only one container in the task, return it
		return task.Containers[0], nil
	}
}

// selectCluster provides the prompt for choosing a cluster
func selectCluster(clusterNames []string) (string, error) {
	if flag.Lookup("test.v") != nil {
		return clusterNames[0], nil
	}
	var clusterName string
	var qs = []*survey.Question{
		{
			Name: "Cluster",
			Prompt: &survey.Select{
				Message: "Select which cluster you want to use:",
				Options: clusterNames,
			},
		},
	}

	err := survey.Ask(qs, &clusterName)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return clusterName, nil
}

// selectService provides the prompt for choosing a service
func selectService(serviceNames []string) (string, error) {
	if flag.Lookup("test.v") != nil {
		return serviceNames[0], nil
	}
	serviceNames = append(serviceNames, "*")
	var serviceName string
	var qs = []*survey.Question{
		{
			Name: "Service",
			Prompt: &survey.Select{
				Message: fmt.Sprintf("Select a service %s:", yellow("(choose * to display all tasks)")),
				Options: serviceNames,
			},
		},
	}

	err := survey.Ask(qs, &serviceName)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return serviceName, nil
}

// selectTask provides the prompt for choosing a Task
func selectTask(tasks []*ecs.Task) (*ecs.Task, error) {
	if flag.Lookup("test.v") != nil {
		return tasks[0], nil
	}
	var options []string
	for _, t := range tasks {
		var containers []string
		for _, c := range t.Containers {
			containers = append(containers, *c.Name)
		}
		id := strings.Split(*t.TaskArn, "/")[2]
		taskDefinion := strings.Split(*t.TaskDefinitionArn, "/")[1]
		options = append(options, fmt.Sprintf("%s\t%s\t(%s)", id, taskDefinion, strings.Join(containers, ",")))
	}

	var qs = []*survey.Question{
		{
			Name: "Task",
			Prompt: &survey.Select{
				Message: fmt.Sprintf("Select the task you would like to connect to %s:", yellow("(if multi-container you will be prompted)")),
				Options: options,
			},
		},
	}

	var selection string
	err := survey.Ask(qs, &selection)
	if err != nil {
		fmt.Println(err.Error())
		return &ecs.Task{}, err
	}

	var task *ecs.Task
	// Loop through our tasks and pull out the one which matches our selection
	for _, t := range tasks {
		id := strings.Split(*t.TaskArn, "/")[2]
		if strings.Contains(selection, id) {
			task = t
			break
		}
	}

	return task, nil
}

// selectContainer prompts the user to choose a container within a task
func selectContainer(containers []*ecs.Container) (*ecs.Container, error) {
	if flag.Lookup("test.v") != nil {
		return containers[0], nil
	}
	var containerNames []string
	for _, c := range containers {
		containerNames = append(containerNames, *c.Name)
	}

	var selection string
	var qs = []*survey.Question{
		{
			Name: "Container",
			Prompt: &survey.Select{
				Message: "More than one container in task, please choose the one you would like to connect to:",
				Options: containerNames,
			},
		},
	}

	err := survey.Ask(qs, &selection)
	if err != nil {
		fmt.Println(err.Error())
		return &ecs.Container{}, err
	}

	var container *ecs.Container
	for _, c := range containers {
		if strings.Contains(*c.Name, selection) {
			container = c
		}
	}

	return container, nil
}

// runCommand executes a command with args, catches any signals and handles them -
// not to be consufed
func runCommand(process string, args ...string) error {
	cmd := exec.Command(process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-sigs:
			}
		}
	}()
	defer close(sigs)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// selectProfile
/* func selectProfile(profiles []string) (string, error) {
	var profile string
	var qs = []*survey.Question{
		{
			Name: "profile",
			Prompt: &survey.Select{
				Message: "Select your AWS Profile",
				Options: profiles,
			},
		},
	}
	// perform the questions
	err := survey.Ask(qs, &profile)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return profile, nil
}
*/

// ReadAwsConfig reads in the aws config file and returns a slice of all profile names
/* func ReadAwsConfig() ([]string, error) {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(fmt.Sprintf("%s/.aws/config", home))
	if err != nil {
		log.Fatal(err)
	}

	var profiles []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Index(line, "[profile ") > -1 {
			raw := strings.Split(line, " ")[1]
			profile := strings.TrimRight(raw, "]")
			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}
*/
