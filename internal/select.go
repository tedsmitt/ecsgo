package app

import (
	"flag"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func init() {
	survey.SelectQuestionTemplate = `
	{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
	{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
	{{- color "default+hb"}}{{ .Message }}{{ .FilterMessage }}{{color "reset"}}
	{{- if .ShowAnswer}}{{color "Cyan"}} {{""}}{{color "reset"}}
	{{- else}}
	  {{- " "}}{{- color "Cyan"}}[Type to filter{{- if and .Help (not .ShowHelp)}}, {{ .Config.HelpInput }} for more help{{end}}]{{color "reset"}}{{- "\n"}}
	  {{- range $ix, $choice := .PageEntries}}
		{{- if eq $ix $.SelectedIndex }}{{color $.Config.Icons.SelectFocus.Format }}{{ $.Config.Icons.SelectFocus.Text }} {{else}}{{color "default"}}  {{end}}
		{{- $choice.Value}}
		{{- color "reset"}}{{"\n"}}
	  {{- end}}
	{{- end}}`
}

// selectCluster provides the prompt for choosing a cluster
func selectCluster(clusterNames []string) (string, error) {
	if flag.Lookup("test.v") != nil {
		return clusterNames[0], nil
	}

	prompt := &survey.Select{
		Message: "Select a cluster:",
		Options: clusterNames,
	}

	var selection string
	err := survey.AskOne(prompt, &selection, survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Text = "➡"
		icons.SelectFocus.Format = "Cyan+hb"
	}))
	if err != nil {
		return "", err
	}

	return selection, nil
}

// selectService provides the prompt for choosing a service
func selectService(serviceNames []string) (string, error) {
	if flag.Lookup("test.v") != nil {
		return serviceNames[0], nil
	}

	serviceNames = append(serviceNames, "*")

	prompt := &survey.Select{
		Message: fmt.Sprintf("Select a service: %s", Yellow("(choose * to display all tasks)")),
		Options: createOpts(serviceNames),
	}

	var selection string
	err := survey.AskOne(prompt, &selection, survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Text = "➡"
		icons.SelectFocus.Format = "Magenta+hb"
	}))
	if err != nil {
		return "", err
	}

	return selection, nil
}

// selectTask provides the prompt for choosing a Task
func selectTask(tasks map[string]*ecs.Task) (*ecs.Task, error) {
	if flag.Lookup("test.v") != nil {
		for _, t := range tasks {
			return t, nil // return the first value from the map
		}
	}

	var taskOpts []string
	for id, t := range tasks {
		taskDefinition := strings.Split(*t.TaskDefinitionArn, "/")[1]
		var containers []string
		for _, c := range t.Containers {
			containers = append(containers, *c.Name)
		}
		taskOpts = append(taskOpts, fmt.Sprintf("%s | %s | (%s)", id, taskDefinition, strings.Join(containers, ",")))
	}

	prompt := &survey.Select{
		Message: "Select a task:",
		Options: createOpts(taskOpts),
	}

	var selection string
	err := survey.AskOne(prompt, &selection, survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Text = "➡"
		icons.SelectFocus.Format = "Green+hb"
	}))
	if err != nil {
		return &ecs.Task{}, err
	}

	if selection == backOpt {
		return &ecs.Task{TaskArn: aws.String(backOpt)}, nil
	}

	taskId := strings.Split(selection, " | ")[0]
	task := tasks[taskId]

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
	var prompt = &survey.Select{
		Message: "Multiple containers found, please select:",
		Options: createOpts(containerNames),
	}

	err := survey.AskOne(prompt, &selection, survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Text = "➡"
		icons.SelectFocus.Format = "Yellow+hb"
	}))
	if err != nil {
		return &ecs.Container{}, err
	}
	if selection == backOpt {
		return &ecs.Container{Name: aws.String(backOpt)}, nil
	}

	var container *ecs.Container
	for _, c := range containers {
		if selection == *c.Name {
			container = c
		}
	}

	return container, nil
}

// inputLocalPort prompts the user to enter a port number for port-forwarding
func inputLocalPort() (string, error) {
	if flag.Lookup("test.v") != nil {
		return "42069", nil
	}

	port := ""
	prompt := &survey.Input{
		Message: "Enter the local port to be used for forwarding\n",
	}
	survey.AskOne(prompt, &port)

	return port, nil
}
