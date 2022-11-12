package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/viper"
)

// executeCommand takes all of our previous values and builds a session for us
// and then calls runCommand to execute the session input via session-manager-plugin
func (e *App) executeCommand() error {
	var command string
	if viper.GetString("cmd") != "" {
		command = viper.GetString("cmd")
	} else {
		if strings.Contains(strings.ToLower(*e.task.PlatformFamily), "windows") {
			command = "powershell.exe"
		} else {
			command = "/bin/sh"
		}
	}
	App, err := e.client.ExecuteCommand(&ecs.ExecuteCommandInput{
		Cluster:     aws.String(e.cluster),
		Interactive: aws.Bool(true),
		Task:        e.task.TaskArn,
		Command:     aws.String(command),
		Container:   e.container.Name,
	})

	if err != nil {
		e.err <- err
		return err
	}

	execSess, err := json.MarshalIndent(App.Session, "", "    ")
	if err != nil {
		e.err <- err
		return err
	}

	taskArnSplit := strings.Split(*e.task.TaskArn, "/")
	taskID := taskArnSplit[len(taskArnSplit)-1]
	target := ssm.StartSessionInput{
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", e.cluster, taskID, *e.container.RuntimeId)),
	}

	targetJson, err := json.MarshalIndent(target, "", "    ")
	if err != nil {
		e.err <- err
		return err
	}

	// Print Cluster/Service/Task information to the console
	fmt.Printf("\nCluster: %v | Service: %v | Task: %s | Cmd: %s", Cyan(e.cluster), Magenta(e.service), Green(strings.Split(*e.task.TaskArn, "/")[2]), Yellow(command))
	fmt.Printf("\nConnecting to container %v\n", Yellow(*e.container.Name))

	// Execute the session-manager-plugin with our task details
	err = runCommand("session-manager-plugin", string(execSess), e.region, "StartSession", "", string(targetJson), e.endpoint)
	e.err <- err

	return err
}
