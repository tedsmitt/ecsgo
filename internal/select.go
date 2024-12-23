/* select.go contains the logic for the Select/Survey views in the TUI app */

package app

import (
	"flag"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/aws/aws-sdk-go-v2/aws"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func createList(items []list.Item, title string) list.Model {
	const defaultWidth = 20
	const listHeight = 14

	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

func runList(l list.Model) (string, error) {
	p := tea.NewProgram(l)
	m, err := p.StartReturningModel()
	if err != nil {
		return "", err
	}

	if l, ok := m.(list.Model); ok {
		if len(l.SelectedItems()) > 0 {
			return l.SelectedItems()[0].(item).title, nil
		}
	}

	return "", fmt.Errorf("no item selected")
}

// createOpts builds the initial options for the survey prompts
func createOpts(opts []string) []string {
	initialOpts := []string{backOpt}
	return append(initialOpts, opts...)
}

// selectCluster provides the prompt for choosing a cluster
func selectCluster(clusterNames []string) (string, error) {
	if flag.Lookup("test.v") != nil {
		if len(clusterNames) > 100 {
			// For Pagination testing, after sorting alphabetically, the 101st cluster is at index 4, and proves
			// that the pagination is working correctly
			return clusterNames[4], nil
		}
		return clusterNames[0], nil
	}

	var items []list.Item
	for _, name := range clusterNames {
		items = append(items, item{title: name})
	}

	l := createList(items, "Select a cluster:")
	selection, err := runList(l)
	if err != nil {
		return "", err
	}

	return selection, nil
}

// selectService provides the prompt for choosing a service
func selectService(serviceNames []string) (string, error) {
	if flag.Lookup("test.v") != nil {
		if len(serviceNames) > int(*awsMaxResults) {
			// For Pagination testing, after sorting alphabetically, the 101st service is at index 4, and proves
			// that the pagination is working correctly
			return serviceNames[4], nil
		}
		return serviceNames[0], nil
	}

	serviceNames = append(serviceNames, "*")

	var items []list.Item
	for _, name := range serviceNames {
		items = append(items, item{title: name})
	}

	l := createList(items, fmt.Sprintf("Select a service: %s", Yellow("(choose * to display all tasks)")))
	selection, err := runList(l)
	if err != nil {
		return "", err
	}

	return selection, nil
}

// selectTask provides the prompt for choosing a Task
func selectTask(tasks map[string]*ecsTypes.Task) (*ecsTypes.Task, error) {
	if flag.Lookup("test.v") != nil {
		// When testing pagination, we want to return a task from the second set of results,
		// which will prove pagination is working correctly
		if len(tasks) > int(*awsMaxResults) {
			return tasks["199"], nil
		}
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

	var items []list.Item
	for _, opt := range taskOpts {
		items = append(items, item{title: opt})
	}

	l := createList(items, "Select a task:")
	selection, err := runList(l)
	if err != nil {
		return &ecsTypes.Task{}, err
	}

	if selection == backOpt {
		return &ecsTypes.Task{TaskArn: aws.String(backOpt)}, nil
	}

	taskId := strings.Split(selection, " | ")[0]
	task := tasks[taskId]

	return task, nil
}

// selectContainer prompts the user to choose a container within a task
func selectContainer(containers *[]ecsTypes.Container) (*ecsTypes.Container, error) {
	if flag.Lookup("test.v") != nil {
		container := *containers
		return &container[0], nil
	}

	var containerNames []string
	for _, c := range *containers {
		containerNames = append(containerNames, *c.Name)
	}

	var items []list.Item
	for _, name := range containerNames {
		items = append(items, item{title: name})
	}

	l := createList(items, "Multiple containers found, please select:")
	selection, err := runList(l)
	if err != nil {
		return &ecsTypes.Container{}, err
	}
	if selection == backOpt {
		return &ecsTypes.Container{Name: aws.String(backOpt)}, nil
	}

	var container *ecsTypes.Container
	for _, c := range *containers {
		cont := c
		if selection == *cont.Name {
			container = &cont
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
