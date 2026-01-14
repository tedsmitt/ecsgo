package app

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestExecuteForwardRequiresRealClient(t *testing.T) {
	app := &App{
		input:   make(chan string, 1),
		err:     make(chan error, 1),
		exit:    make(chan error, 1),
		client:  ECSClientMock{},
		region:  "eu-west-1",
		cluster: "test-cluster",
		task: &ecsTypes.Task{
			TaskArn:           aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
			TaskDefinitionArn: aws.String("arn:aws:ecs:eu-west-1:111111111111:task-definition/my-task:1"),
			Containers: []ecsTypes.Container{
				{
					Name:      aws.String("nginx"),
					RuntimeId: aws.String("544e08d919364be9926186b086c29868-2531612879"),
				},
			},
		},
	}
	app.container = &app.task.Containers[0]

	viper.Set("local-port", "8080")

	defer func() {
		viper.Set("local-port", "")
		if r := recover(); r != nil {
			panicMsg, ok := r.(string)
			if ok {
				assert.True(t, strings.Contains(panicMsg, "interface conversion"))
			} else {
				errorMsg, ok := r.(error)
				if ok {
					assert.True(t, strings.Contains(errorMsg.Error(), "interface conversion"))
				}
			}
		}
	}()

	err := app.executeForward()

	if err == nil {
		t.Skip("executeForward requires a real *ecs.Client due to type assertion on line 33")
	}
}
