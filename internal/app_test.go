package app

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("AWS_DEFAULT_REGION", "eu-west-1")
}

type ECSClientMock struct {
	ListClustersMock               func(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error)
	ListServicesMock               func(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error)
	ListTasksMock                  func(ctx context.Context, params *ecs.ListTasksInput, optFns ...func(*ecs.Options)) (*ecs.ListTasksOutput, error)
	DescribeTasksMock              func(ctx context.Context, params *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error)
	DescribeTaskDefinitionMock     func(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error)
	DescribeContainerInstancesMock func(ctx context.Context, params *ecs.DescribeContainerInstancesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeContainerInstancesOutput, error)
	ExecuteCommandMock             func(ctx context.Context, params *ecs.ExecuteCommandInput, optFns ...func(*ecs.Options)) (*ecs.ExecuteCommandOutput, error)
}

func (m ECSClientMock) ListClusters(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
	return m.ListClustersMock(ctx, params, optFns...)
}

func (m ECSClientMock) ListServices(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
	return m.ListServicesMock(ctx, params, optFns...)
}

func (m ECSClientMock) ListTasks(ctx context.Context, params *ecs.ListTasksInput, optFns ...func(*ecs.Options)) (*ecs.ListTasksOutput, error) {
	return m.ListTasksMock(ctx, params, optFns...)
}

func (m ECSClientMock) DescribeTasks(ctx context.Context, params *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error) {
	return m.DescribeTasksMock(ctx, params, optFns...)
}

func (m ECSClientMock) DescribeTaskDefinition(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
	return m.DescribeTaskDefinitionMock(ctx, params, optFns...)
}

func (m ECSClientMock) DescribeContainerInstances(ctx context.Context, params *ecs.DescribeContainerInstancesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeContainerInstancesOutput, error) {
	return m.DescribeContainerInstancesMock(ctx, params, optFns...)
}
func (m ECSClientMock) ExecuteCommand(ctx context.Context, params *ecs.ExecuteCommandInput, optFns ...func(*ecs.Options)) (*ecs.ExecuteCommandOutput, error) {
	return m.ExecuteCommandMock(ctx, params, optFns...)
}

type MockEC2API struct {
	ec2.Client            // embedding of the interface is needed to skip implementation of all methods
	DescribeInstancesMock func(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

func (m *MockEC2API) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	if m.DescribeInstancesMock != nil {
		return m.DescribeInstancesMock(input)
	}
	return nil, nil
}

// CreateMockApp initialises a new App struct and takes a MockClient as an argument - only used in tests
func CreateMockApp(c ECSClient) *App {
	e := &App{
		input:  make(chan string, 1),
		err:    make(chan error, 1),
		exit:   make(chan error, 1),
		client: c,
		region: "eu-west-1",
	}

	return e
}

func TestGetCluster(t *testing.T) {
	paginationCall := 0
	cases := []struct {
		name     string
		client   func(t *testing.T) ECSClient
		expected string
	}{
		{
			name: "TestGetClusterWithResultsPaginated",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					ListClustersMock: func(ctx context.Context, input *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
						var clusters []string
						for i := paginationCall; i < (paginationCall * 100); i++ {
							clusters = append(clusters, *aws.String(fmt.Sprintf("arn:aws:ecs:eu-west-1:1111111111:cluster/test-cluster-%d", i)))
						}
						paginationCall = paginationCall + 1
						if paginationCall > 2 {
							return &ecs.ListClustersOutput{
								ClusterArns: clusters,
								NextToken:   nil,
							}, nil
						}
						return &ecs.ListClustersOutput{
							ClusterArns: clusters,
							NextToken:   aws.String("test-token"),
						}, nil
					},
				}
			},
			expected: "test-cluster-101",
		},
	}

	for _, c := range cases {
		client := c.client(t)
		input := CreateMockApp(client)
		input.getCluster()
		if ok := assert.Equal(t, c.expected, input.cluster); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}

func TestGetService(t *testing.T) {
	paginationCall := 1
	cases := []struct {
		name     string
		client   func(t *testing.T) ECSClient
		expected string
	}{
		{
			name: "TestGetServiceWithResults",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					ListServicesMock: func(ctx context.Context, input *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
						return &ecs.ListServicesOutput{
							ServiceArns: []string{
								*aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/App/test-service-1"),
								*aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/blueGreen/test-service-2"),
							},
						}, nil
					},
				}
			},
			expected: "test-service-1",
		},
		{
			name: "TestGetServiceWithResultsPaginated",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					ListServicesMock: func(ctx context.Context, input *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
						var services []string
						for i := paginationCall; i < (paginationCall * 100); i++ {
							services = append(services, *aws.String(fmt.Sprintf("arn:aws:ecs:eu-west-1:1111111111:cluster/App/test-service-%d", i)))
						}
						paginationCall = paginationCall + 1
						if paginationCall > 2 {
							return &ecs.ListServicesOutput{
								ServiceArns: services,
								NextToken:   nil,
							}, nil
						}
						return &ecs.ListServicesOutput{
							ServiceArns: services,
							NextToken:   aws.String("test-token"),
						}, nil
					},
				}
			},
			expected: "test-service-101",
		},
		{
			name: "TestGetServiceWithoutResults",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					ListServicesMock: func(ctx context.Context, input *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
						return &ecs.ListServicesOutput{
							ServiceArns: []string{},
						}, nil
					},
				}
			},
			expected: "",
		},
	}

	for _, c := range cases {
		client := c.client(t)
		input := CreateMockApp(client)
		input.cluster = "App"
		input.getService()
		if ok := assert.Equal(t, c.expected, input.service); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}

func TestGetTask(t *testing.T) {
	paginationCall := 1
	cases := []struct {
		name     string
		client   func(t *testing.T) ECSClient
		expected *ecsTypes.Task
	}{
		{
			name: "TestGetTaskWithResults",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					ListTasksMock: func(ctx context.Context, input *ecs.ListTasksInput, optFns ...func(*ecs.Options)) (*ecs.ListTasksOutput, error) {
						return &ecs.ListTasksOutput{
							TaskArns: []string{
								*aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
							},
						}, nil
					},
					DescribeTasksMock: func(ctx context.Context, input *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error) {
						var tasks []ecsTypes.Task
						for _, taskArn := range input.Tasks {
							tasks = append(tasks, ecsTypes.Task{TaskArn: &taskArn, LaunchType: ecsTypes.LaunchTypeFargate})
						}
						return &ecs.DescribeTasksOutput{
							Tasks: tasks,
						}, nil
					},
				}
			},
			expected: &ecsTypes.Task{
				TaskArn:    aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				LaunchType: ecsTypes.LaunchTypeFargate,
			},
		},
		{
			name: "TestGetTaskWithResultsPaginated",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					ListTasksMock: func(ctx context.Context, input *ecs.ListTasksInput, optFns ...func(*ecs.Options)) (*ecs.ListTasksOutput, error) {
						var taskArns []string
						var tasks []*ecsTypes.Task
						for i := paginationCall; i < (paginationCall * 100); i++ {
							taskArn := *aws.String(fmt.Sprintf("arn:aws:ecs:eu-west-1:111111111111:task/App/%d", i))
							taskArns = append(taskArns, taskArn)
							tasks = append(tasks, &ecsTypes.Task{TaskArn: &taskArn, LaunchType: ecsTypes.LaunchTypeFargate})
						}
						paginationCall = paginationCall + 1
						if paginationCall > 2 {
							return &ecs.ListTasksOutput{
								TaskArns:  taskArns,
								NextToken: nil,
							}, nil
						}
						return &ecs.ListTasksOutput{
							TaskArns:  taskArns,
							NextToken: aws.String("test-token"),
						}, nil
					},
					DescribeTasksMock: func(ctx context.Context, input *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error) {
						var tasks []ecsTypes.Task
						for _, taskArn := range input.Tasks {
							tasks = append(tasks, ecsTypes.Task{TaskArn: &taskArn, LaunchType: ecsTypes.LaunchTypeFargate})
						}
						return &ecs.DescribeTasksOutput{
							Tasks: tasks,
						}, nil
					},
				}
			},
			expected: &ecsTypes.Task{
				TaskArn:    aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/199"),
				LaunchType: ecsTypes.LaunchTypeFargate,
			},
		},
		{
			name: "TestGetTaskWithoutResults",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					ListTasksMock: func(ctx context.Context, input *ecs.ListTasksInput, optFns ...func(*ecs.Options)) (*ecs.ListTasksOutput, error) {
						return &ecs.ListTasksOutput{
							TaskArns: []string{},
						}, nil
					},
				}
			},
			expected: nil,
		},
	}

	for _, c := range cases {
		client := c.client(t)
		input := CreateMockApp(client)
		input.cluster = "App"
		input.service = "test-service-1"
		input.getTask()
		if ok := assert.Equal(t, c.expected, input.task); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}

func TestGetContainer(t *testing.T) {
	cases := []struct {
		name     string
		client   func(t *testing.T) ECSClient
		task     *ecsTypes.Task
		expected *ecsTypes.Container
	}{
		{
			name: "TestGetContainerWithMultipleContainers",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{}
			},
			task: &ecsTypes.Task{
				Containers: []ecsTypes.Container{
					{
						Name: aws.String("echo-server"),
					},
					{
						Name: aws.String("redis"),
					},
				},
			},
			expected: &ecsTypes.Container{
				Name: aws.String("echo-server"),
			},
		},
		{
			name: "TestGetContainerWithSingleContainer",
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{}
			},
			task: &ecsTypes.Task{
				Containers: []ecsTypes.Container{
					{
						Name: aws.String("nginx"),
					},
				},
			},
			expected: &ecsTypes.Container{
				Name: aws.String("nginx"),
			},
		},
	}

	for _, c := range cases {
		client := c.client(t)
		input := CreateMockApp(client)
		input.task = c.task
		input.getContainer()
		if ok := assert.Equal(t, c.expected, input.container); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}
