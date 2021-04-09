package cli

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("AWS_DEFAULT_REGION", "eu-west-1")
}

type MockECSAPI struct {
	ecsiface.ECSAPI    // embedding of the interface is needed to skip implementation of all methods
	ListClustersMock   func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
	ListServicesMock   func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error)
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

func (m *MockECSAPI) ListServices(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) { // This allows the test to use the same method
	if m.ListServicesMock != nil {
		return m.ListServicesMock(input) // We intercept and return a made up reply
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

func (m *MockECSAPI) ExecuteCommand(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) { // This allows the test to use the same method
	if m.ExecuteCommandMock != nil {
		return m.ExecuteCommandMock(input) // We intercept and return a made up reply
	}
	return nil, nil // return any value you think is good for you
}

// CreateMockExecCommand initialises a new ExecCommand struct and takes a MockClient
// as an argument - only used in unit tests.
func CreateMockExecCommand(c *MockECSAPI) *ExecCommand {
	e := &ExecCommand{
		cmd:      make(chan string, 1),
		err:      make(chan error, 1),
		done:     make(chan bool),
		client:   c,
		region:   "eu-west-1",
		endpoint: "ecs.eu-west-1.amazonaws.com",
	}
	return e
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
		cmd := CreateMockExecCommand(c.client)
		cmd.getCluster()
		assert.Equal(t, c.expected, cmd.cluster)
	}
}

func TestGetService(t *testing.T) {
	cases := []struct {
		name     string
		client   *MockECSAPI
		expected string
	}{
		{
			name: "TestGetServiceWithResults",
			client: &MockECSAPI{
				ListServicesMock: func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
					return &ecs.ListServicesOutput{
						ServiceArns: []*string{
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/execCommand/test-service-1"),
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/bluegreen/test-service-2"),
						},
					}, nil
				},
			},
			expected: "test-service-1",
		},
		{
			name: "TestGetServiceWithoutResults",
			client: &MockECSAPI{
				ListServicesMock: func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
					return &ecs.ListServicesOutput{
						ServiceArns: []*string{},
					}, nil
				},
			},
			expected: "",
		},
	}
	for _, c := range cases {
		cmd := CreateMockExecCommand(c.client)
		cmd.cluster = "execCommand"
		cmd.getService()
		assert.Equal(t, c.expected, cmd.service)
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
			expected: nil,
		},
	}
	for _, c := range cases {
		cmd := CreateMockExecCommand(c.client)
		cmd.cluster = "execCommand"
		cmd.service = "test-service-1"
		cmd.getTask()
		assert.Equal(t, c.expected, cmd.task)
	}
}

func TestGetContainer(t *testing.T) {
	cases := []struct {
		name     string
		client   *MockECSAPI
		task     *ecs.Task
		expected *ecs.Container
	}{
		{
			name:   "TestGetContainerWithMultipleContainers",
			client: &MockECSAPI{},
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
			name:   "TestGetContainerWithSingleContainer",
			client: &MockECSAPI{},
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
		cmd := CreateMockExecCommand(c.client)
		cmd.task = c.task
		cmd.getContainer()
		assert.Equal(t, c.expected, cmd.container)
	}
}

/* func TestStartExecuteCommand(t *testing.T) {
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
*/
