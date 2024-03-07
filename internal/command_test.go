package app

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/assert"
)

func TestExecuteInput(t *testing.T) {
	cases := []struct {
		name     string
		expected error
		client   *MockECSAPI
		cluster  string
		task     *ecsTypes.Task
	}{
		{
			name:    "TestExecuteInput",
			cluster: "test",
			task: &ecsTypes.Task{
				TaskArn: aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				Containers: []ecsTypes.Container{
					{
						Name:      aws.String("nginx"),
						RuntimeId: aws.String("544e08d919364be9926186b086c29868-2531612879"),
					},
				},
				PlatformFamily: aws.String("Linux"),
			},
			ecsClient: &MockECSAPI{
				ExecuteCommandMock: func(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) {
					return &ecs.ExecuteCommandOutput{
						Session: &ecsTypes.Session{
							SessionId:  aws.String("ecs-execute-command-0e86561fddf625dc1"),
							StreamUrl:  aws.String("wss://ssmmessages.eu-west-1.amazonaws.com/v1/data-channel/ecs-execute-command-blah"),
							TokenValue: aws.String("abc123"),
						},
					}, nil
				},
			},
			expected: nil,
		},
	}

	for _, c := range cases {
		app := &App{
			input:    make(chan string, 1),
			err:      make(chan error, 1),
			exit:     make(chan error, 1),
			client:   &c.client.Client,
			region:   "eu-west-1",
			endpoint: "ecs.eu-west-1.amazonaws.com",
			cluster:  c.cluster,
			task:     *c.task,
		}
		app.container = &c.task.Containers[0]
		err := app.executeCommand()
		if ok := assert.Equal(t, c.expected, err); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}
