package cmd

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func StartExecuteCommand() {

	svc := createEcsClient()

	/* 	resp, err := svc.ListClusters(&ecs.ListClustersInput{})
	   	if err != nil {
	   		log.Println(err)
	   	}
	   	task, _ := svc.ListTasks(&ecs.ListTasksInput{
	   		Cluster: aws.String(clusterName),
	   	}) */
	taskId := "808836463e444bae8e476a108dad75b5"
	runtimeId := "808836463e444bae8e476a108dad75b5-1527056392"

	input := &ecs.ExecuteCommandInput{
		Cluster:     aws.String("execCommand"),
		Interactive: aws.Bool(true),
		Task:        aws.String(taskId),
		Command:     aws.String("/bin/sh"),
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*15)
	execCmd, err := svc.ExecuteCommandWithContext(ctx, input)
	if err != nil {
		panic(err)
	}

	sessJson, err := json.Marshal(execCmd.Session)
	if err != nil {
		panic(err)
	}

	paramsInput := ssm.StartSessionInput{
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", "execCommand", taskId, runtimeId)),
	}
	paramsJson, err := json.Marshal(paramsInput)
	if err != nil {
		log.Println(err)
	}

	endpoint := "https://ecs.eu-west-1.amazonaws.com"

	call := exec.Command("session-manager-plugin", string(sessJson),
		"eu-west-1", "StartSession", "edintheclouds-dev", string(paramsJson), endpoint)
	call.Stderr = os.Stderr
	call.Stdout = os.Stdout
	call.Stdin = os.Stdin

	// ignore signal(sigint)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-sigs:
			case <-done:
				break
			}
		}
	}()
	defer close(done)

	// run subprocess
	if err := call.Run(); err != nil {
		log.Println("here")
		log.Println(err)
	}
	return
}
