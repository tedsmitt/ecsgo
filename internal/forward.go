package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/viper"
)

// executeInput takes all of our previous values and builds a session for us
// and then calls runCommand to execute the session input via session-manager-plugin
func (e *App) executeForward() error {
	taskArnSplit := strings.Split(*e.task.TaskArn, "/")
	taskID := taskArnSplit[len(taskArnSplit)-1]
	target := ssm.StartSessionInput{
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", e.cluster, taskID, *e.container.RuntimeId)),
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(viper.GetString("profile")),
		config.WithRegion(region),
	)
	if err != nil {
		panic(err)
	}
	client := ssm.NewFromConfig(cfg) // TODO: add region
	containerPort, err := getContainerPort(e.client, *e.task.TaskDefinitionArn, *e.container.Name)
	if err != nil {
		e.err <- err
		return err
	}
	localPort := viper.GetString("local-port")
	if localPort == "" {
		localPort, err = inputLocalPort()
		if err != nil {
			e.err <- err
			return err
		}
	}
	portNumber := fmt.Sprint(*containerPort)
	input := &ssm.StartSessionInput{
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters: map[string][]string{
			"localPortNumber": {localPort},
			"portNumber":      {portNumber},
		},
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", e.cluster, taskID, *e.container.RuntimeId)),
	}
	sess, err := client.StartSession(context.TODO(), input)
	if err != nil {
		e.err <- err
		return err
	}
	sessJson, err := json.Marshal(sess)
	if err != nil {
		e.err <- err
		return err
	}
	paramsJson, err := json.Marshal(target)
	if err != nil {
		e.err <- err
		return err
	}

	// Print Cluster/Service/Task information to the console
	if !viper.GetBool("quiet") {
		fmt.Printf("\nCluster: %v | Service: %v | Task: %s", Cyan(e.cluster), Magenta(e.service), Green(strings.Split(*e.task.TaskArn, "/")[2]))
		fmt.Printf("\nPort-forwarding %s:%d -> container %v\n", localPort, *containerPort, Yellow(*e.container.Name))
	}

	// Execute the session-manager-plugin with our task details
	err = runCommand("session-manager-plugin", string(sessJson), e.region, "StartSession", "", string(paramsJson))
	e.err <- err
	return err

}
