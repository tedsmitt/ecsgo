package cmd

import (
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
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var (
	region   string
	endpoint string

	red     = color.New(color.FgRed).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()

	backOpt = "⏎ Back"

	version = "unset"
	commit  = "unset"
	date    = "unset"
	builtBy = "unset"
)

func init() {
	survey.SelectQuestionTemplate = `
	{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
	{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
	{{- color "default+hb"}}{{ .Message }}{{ .FilterMessage }}{{color "reset"}}
	{{- if .ShowAnswer}}{{color "cyan"}} {{""}}{{color "reset"}}
	{{- else}}
	  {{- " "}}{{- color "cyan"}}[Type to filter{{- if and .Help (not .ShowHelp)}}, {{ .Config.HelpInput }} for more help{{end}}]{{color "reset"}}{{- "\n"}}
	  {{- range $ix, $choice := .PageEntries}}
		{{- if eq $ix $.SelectedIndex }}{{color $.Config.Icons.SelectFocus.Format }}{{ $.Config.Icons.SelectFocus.Text }} {{else}}{{color "default"}}  {{end}}
		{{- $choice.Value}}
		{{- color "reset"}}{{"\n"}}
	  {{- end}}
	{{- end}}`
}

func createEcsClient() *ecs.ECS {
	region := viper.GetString("region")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		Profile:           viper.GetString("profile"),
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ecs.New(sess)

	return client
}

func createOpts(opts []string) []string {
	initialOpts := []string{backOpt}
	return append(initialOpts, opts...)
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
		icons.SelectFocus.Format = "cyan+hb"
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
		Message: fmt.Sprintf("Select a service: %s", yellow("(choose * to display all tasks)")),
		Options: createOpts(serviceNames),
	}

	var selection string
	err := survey.AskOne(prompt, &selection, survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Text = "➡"
		icons.SelectFocus.Format = "magenta+hb"
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
		icons.SelectFocus.Format = "green+hb"
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
		icons.SelectFocus.Format = "yellow+hb"
	}))
	if err != nil {
		return &ecs.Container{}, err
	}
	if selection == backOpt {
		return &ecs.Container{Name: aws.String(backOpt)}, nil
	}

	var container *ecs.Container
	for _, c := range containers {
		if strings.Contains(*c.Name, selection) {
			container = c
		}
	}

	return container, nil
}

// runCommand executes a command with args
func runCommand(process string, args ...string) error {
	cmd := exec.Command(process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	// Capture any SIGINTs and discard them
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT)
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

// getVersion returns version information
func getVersion() string {
	return fmt.Sprintf("Version: %s, Commit: %s, Built date: %s, Built by: %s", version, commit, date, builtBy)
}
