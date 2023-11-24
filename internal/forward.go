package app

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
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

	mySession := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		Profile:           viper.GetString("profile"),
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ssm.New(mySession)
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
	input := &ssm.StartSessionInput{
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters: map[string][]*string{
			"portNumber":      {aws.String(strconv.FormatInt(*containerPort, 10))},
			"localPortNumber": {aws.String(localPort)},
		},
		Target: aws.String(fmt.Sprintf("ecs:%s_%s_%s", e.cluster, taskID, *e.container.RuntimeId)),
	}
	sess, err := client.StartSession(input)
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
	fmt.Printf("\nCluster: %v | Service: %v | Task: %s", Cyan(e.cluster), Magenta(e.service), Green(strings.Split(*e.task.TaskArn, "/")[2]))
	fmt.Printf("\nPort-forwarding %s:%d -> container %v\n", localPort, *containerPort, Yellow(*e.container.Name))

	// Execute the session-manager-plugin with our task details
	err = runCommand("session-manager-plugin", string(sessJson), e.region, "StartSession", "", string(paramsJson))
	e.err <- err
	return err

}
