package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func StartExecuteCommand() {

	client := createEcsClient()

	clusterName, err := getCluster(client)
	if err != nil {
		log.Fatalf(red(err))
	}
	task, err := getTask(client, clusterName)
	if err != nil {
		log.Fatalf(red(err))
	}
	container, err := getContainer(client, task)
	if err != nil {
		log.Fatalf(red(err))
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
		log.Fatal(err)
	}
	execSess, err := json.Marshal(execCommand.Session)
	if err != nil {
		log.Fatal(err)
	}
	target := ssm.StartSessionInput{
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", "execCommand", *task.TaskArn, *container.RuntimeId)),
	}
	targetJson, err := json.Marshal(target)
	if err != nil {
		log.Println(err)
	}

	// Expecting session-manager-plugin to be found in $PATH
	runCommand("session-manager-plugin", string(execSess),
		region, "StartSession", "", string(targetJson), endpoint)
	if err != nil {
		log.Fatalf(err.Error())
	}
	os.Exit(0)
}
