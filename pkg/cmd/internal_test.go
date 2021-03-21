package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/stretchr/testify/assert"
)

type MockECSAPI struct {
	ecsiface.ECSAPI    // embedding of the interface is needed to skip implementation of all methods
	ListClustersMock   func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
	ListTasksMock      func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error)
	DescribeTasksMock  func(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error)
	ExecuteCommandMock func(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error)
}

func (m *MockECSAPI) ListClusters(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) { // This allows the test to use the same method
	if m.ListClustersMock != nil {
		return m.ListClustersMock(input) // We intercept and return a made up reply
	}
	return nil, nil // return any value you think is good for you
}

func (m *MockECSAPI) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) { // This allows the test to use the same method
	if m.ListTasksMock != nil {
		return m.ListTasksMock(input) // We intercept and return a made up reply
	}
	return nil, nil // return any value you think is good for you
}

func (m *MockECSAPI) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) { // This allows the test to use the same method
	if m.DescribeTasksMock != nil {
		return m.DescribeTasksMock(input) // We intercept and return a made up reply
	}
	return nil, nil // return any value you think is good for you
}

func TestGetCluster(t *testing.T) {
	cases := []struct {
		name     string
		client   *MockECSAPI
		expected string
	}{
		{
			name: "TestGetClusterWithResults",
			client: &MockECSAPI{
				ListClustersMock: func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
					return &ecs.ListClustersOutput{
						ClusterArns: []*string{
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/execCommand"),
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/bluegreen"),
						},
					}, nil
				},
			},
			expected: "execCommand",
		},
		{
			name: "TestGetClusterWithoutResults",
			client: &MockECSAPI{
				ListClustersMock: func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
					return &ecs.ListClustersOutput{
						ClusterArns: []*string{},
					}, nil
				},
			},
			expected: "",
		},
	}
	for _, c := range cases {
		result, _ := getCluster(c.client)
		assert.Equal(t, c.expected, result)
	}
}

func TestGetTask(t *testing.T) {
	cases := []struct {
		name     string
		client   *MockECSAPI
		expected *ecs.Task
	}{
		{
			name: "TestGetTaskWithResults",
			client: &MockECSAPI{
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
							},
						},
					}, nil
				},
			},
			expected: &ecs.Task{
				TaskArn: aws.String("arn:aws:ecs:eu-west-1:111111111111:task/execCommand/8a58117dac38436ba5547e9da5d3ac3d"),
			},
		},
		{
			name: "TestGetTaskWithoutResults",
			client: &MockECSAPI{
				ListTasksMock: func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
					return &ecs.ListTasksOutput{
						TaskArns: []*string{},
					}, nil
				},
			},
			expected: &ecs.Task{},
		},
	}
	for _, c := range cases {
		clusterName := "execCommand"
		result, _ := getTask(c.client, clusterName)
		assert.Equal(t, c.expected, result)
	}
}

func TestGetContainer(t *testing.T) {
	cases := []struct {
		name     string
		task     *ecs.Task
		expected *ecs.Container
	}{
		{
			name: "TestGetContainerWithMultipleContainers",
			task: &ecs.Task{
				Containers: []*ecs.Container{
					{
						Name: aws.String("echo-server"),
					},
					{
						Name: aws.String("redis"),
					},
				},
			},
			expected: &ecs.Container{
				Name: aws.String("echo-server"),
			},
		},
		{
			name: "TestGetContainerWithSingleContainer",
			task: &ecs.Task{
				Containers: []*ecs.Container{
					{
						Name: aws.String("nginx"),
					},
				},
			},
			expected: &ecs.Container{
				Name: aws.String("nginx"),
			},
		},
	}
	for _, c := range cases {
		result, _ := getContainer(c.task)
		assert.Equal(t, c.expected, result)
	}
}
