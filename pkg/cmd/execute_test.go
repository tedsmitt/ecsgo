package cmd

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("AWS_DEFAULT_REGION", "eu-west-1")
}

func (m *MockECSAPI) ExecuteCommand(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) { // This allows the test to use the same method
	if m.ExecuteCommandMock != nil {
		return m.ExecuteCommandMock(input) // We intercept and return a made up reply
	}
	return nil, nil // return any value you think is good for you
}

func TestStartExecuteCommand(t *testing.T) {
	cases := []struct {
		name     string
		client   *MockECSAPI
		expected error
	}{
		{
			name: "TestStartExecuteCommandWithClusters",
			client: &MockECSAPI{
				ListClustersMock: func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
					return &ecs.ListClustersOutput{
						ClusterArns: []*string{
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/execCommand"),
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/bluegreen"),
						},
					}, nil
				},
				ListServicesMock: func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
					return &ecs.ListServicesOutput{
						ServiceArns: []*string{
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/execCommand/test-service-1"),
						},
					}, nil
				},
				ListTasksMock: func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
					return &ecs.ListTasksOutput{
						TaskArns: []*string{
							aws.String("arn:aws:ecs:eu-west-1:111111111111:task/execCommand/8a58117dac38436ba5547e9da5d3ac3d"),
						},
					}, nil
				},
				DescribeTasksMock: func(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
					return &ecs.DescribeTasksOutput{
						Tasks: []*ecs.Task{
							{
								TaskArn: aws.String("arn:aws:ecs:eu-west-1:111111111111:task/execCommand/8a58117dac38436ba5547e9da5d3ac3d"),
								Containers: []*ecs.Container{
									{
										Name:      aws.String("echo-server"),
										RuntimeId: aws.String("8a58117dac38436ba5547e9da5d3ac3d-1527056392"),
									},
								},
							},
						},
					}, nil
				},
				ExecuteCommandMock: func(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) {
					return &ecs.ExecuteCommandOutput{
						Session: &ecs.Session{
							SessionId:  aws.String("ecs-execute-command-05b8e510e3433762c"),
							StreamUrl:  aws.String("wss://ssmmessages.eu-west-1.amazonaws.com/v1/data-channel/ecs-execute-command-05b8e510e3433762c?role=publish_subscribe"),
							TokenValue: aws.String("ABCDEF123456"),
						},
					}, nil
				},
			},
			expected: nil, // If we execute with the session details above then we actually get a clean exit from session-manager-plugin, so we don't expect an error
		},
	}
	for _, c := range cases {
		result := StartExecuteCommand(c.client)
		assert.Equal(t, c.expected, result)
	}
}
