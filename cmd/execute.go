package cmd

import (
	_ "embed"
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

	clusterName, err := getCluster()
	if err != nil {
		log.Fatal(err)
	}
	task, err := getTask(clusterName)
	if err != nil {
		log.Fatal(err)
	}
	container, err := getContainer(task)
	if err != nil {
		log.Fatal(err)
	}
	// Check if command has been passed to the tool, otherwise default
	// to /bin/sh
	var command string
	if viper.GetString("cmd") != "" {
		command = viper.GetString("cmd")
	} else {
		command = "/bin/sh"
	}

	// construct our execute command
	input := &ecs.ExecuteCommandInput{
		Cluster:     aws.String(clusterName),
		Interactive: aws.Bool(true),
		Task:        task.TaskArn,
		Command:     aws.String(command),
		Container:   container.Name,
	}
	//ctx, _ := context.WithTimeout(context.Background(), time.Second*15)
	execCommand, err := client.ExecuteCommand(input)
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

	runCommand("session-manager-plugin", string(execSess),
		region, "StartSession", "", string(targetJson), endpoint)
	if err != nil {
		log.Fatalf(err.Error())
	}
	os.Exit(0)
}
