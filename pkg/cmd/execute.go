package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func StartExecuteCommand(client ecsiface.ECSAPI) error {

	clusterName, err := getCluster(client)
	if err != nil {
		return err
	}
	task, err := getTask(client, clusterName)
	if err != nil {
		return err
	}
	container, err := getContainer(task)
	if err != nil {
		return err
	}

	// Check if command has been passed to the tool, otherwise default
	// to /bin/sh
	var command string
	if viper.GetString("cmd") != "" {
		command = viper.GetString("cmd")
	} else {
		command = "/bin/sh"
	}

	execCommand, err := client.ExecuteCommand(&ecs.ExecuteCommandInput{
		Cluster:     aws.String(clusterName),
		Interactive: aws.Bool(true),
		Task:        task.TaskArn,
		Command:     aws.String(command),
		Container:   container.Name,
	})
	if err != nil {
		return err
	}
	execSess, err := json.Marshal(execCommand.Session)
	if err != nil {
		return err
	}
	target := ssm.StartSessionInput{
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", clusterName, *task.TaskArn, *container.RuntimeId)),
	}
	targetJson, err := json.Marshal(target)
	if err != nil {
		return err
	}

	// Expecting session-manager-plugin to be found in $PATH
	if err = runCommand("session-manager-plugin", string(execSess), region, "StartSession", "", string(targetJson), endpoint); err != nil {
		return err
	}
	return nil
}
