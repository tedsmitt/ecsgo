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

type MockECSAPI struct {
	ecs.Client                     // embedding of the interface is needed to skip implementation of all methods
	ListClustersMock               func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
	ListServicesMock               func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error)
	ListTasksMock                  func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error)
	DescribeTasksMock              func(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error)
	DescribeTaskDefinitionMock     func(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error)
	DescribeContainerInstancesMock func(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
	ExecuteCommandMock             func(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error)
}

func (m *MockECSAPI) ListClusters(ctx context.Context, input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	if m.ListClustersMock != nil {
		return m.ListClustersMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) ListServices(ctx context.Context, input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
	if m.ListServicesMock != nil {
		return m.ListServicesMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) ListTasks(ctx context.Context, input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	if m.ListTasksMock != nil {
		return m.ListTasksMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) DescribeTasks(ctx context.Context, input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	if m.DescribeTasksMock != nil {
		return m.DescribeTasksMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) DescribeTaskDefinition(ctx context.Context, input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	if m.DescribeTaskDefinitionMock != nil {
		return m.DescribeTaskDefinitionMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) DescribeContainerInstances(ctx context.Context, input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	if m.DescribeContainerInstancesMock != nil {
		return m.DescribeContainerInstancesMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) ExecuteCommand(ctx context.Context, input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) {
	if m.ExecuteCommandMock != nil {
		return m.ExecuteCommandMock(input)
	}
	return nil, nil
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
func CreateMockApp(c *MockECSAPI) *App {
	e := &App{
		input:    make(chan string, 1),
		err:      make(chan error, 1),
		exit:     make(chan error, 1),
		client:   &c.Client,
		region:   "eu-west-1",
		endpoint: "ecs.eu-west-1.amazonaws.com",
	}

	return e
}

func TestGetCluster(t *testing.T) {
	paginationCall := 0
	cases := []struct {
		name     string
		client   MockECSAPI
		expected string
	}{
		{
			name: "TestGetClusterWithResultsPaginated",
			client: MockECSAPI{
				ListClustersMock: func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
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
			},
			expected: "test-cluster-101",
		},
	}

	for _, c := range cases {
		input := CreateMockApp(&c.client)
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
		name      string
		ecsClient *MockECSAPI
		expected  string
	}{
		{
			name: "TestGetServiceWithResults",
			ecsClient: &MockECSAPI{
				ListServicesMock: func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
					return &ecs.ListServicesOutput{
						ServiceArns: []string{
							*aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/App/test-service-1"),
							*aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/blueGreen/test-service-2"),
						},
					}, nil
				},
			},
			expected: "test-service-1",
		},
		{
			name: "TestGetServiceWithResultsPaginated",
			ecsClient: &MockECSAPI{
				ListServicesMock: func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
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
			},
			expected: "test-service-101",
		},
		{
			name: "TestGetServiceWithoutResults",
			ecsClient: &MockECSAPI{
				ListServicesMock: func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
					return &ecs.ListServicesOutput{
						ServiceArns: []string{},
					}, nil
				},
			},
			expected: "",
		},
	}

	for _, c := range cases {
		input := CreateMockApp(c.ecsClient)
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
		client   *MockECSAPI
		expected *ecsTypes.Task
	}{
		{
			name: "TestGetTaskWithResults",
			ecsClient: &MockECSAPI{
				ListTasksMock: func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
					return &ecs.ListTasksOutput{
						TaskArns: []string{
							*aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
						},
					}, nil
				},
				DescribeTasksMock: func(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
					var tasks []ecsTypes.Task
					for _, taskArn := range input.Tasks {
						tasks = append(tasks, ecsTypes.Task{TaskArn: &taskArn, LaunchType: ecsTypes.LaunchTypeFargate})
					}
					return &ecs.DescribeTasksOutput{
						Tasks: tasks,
					}, nil
				},
			},
			expected: &ecsTypes.Task{
				TaskArn:    aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				LaunchType: ecsTypes.LaunchTypeFargate,
			},
		},
		{
			name: "TestGetTaskWithResultsPaginated",
			ecsClient: &MockECSAPI{
				ListTasksMock: func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
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
				DescribeTasksMock: func(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
					var tasks []ecsTypes.Task
					for _, taskArn := range input.Tasks {
						tasks = append(tasks, ecsTypes.Task{TaskArn: &taskArn, LaunchType: ecsTypes.LaunchTypeFargate})
					}
					return &ecs.DescribeTasksOutput{
						Tasks: tasks,
					}, nil
				},
			},
			expected: &ecsTypes.Task{
				TaskArn:    aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/199"),
				LaunchType: ecsTypes.LaunchTypeFargate,
			},
		},
		{
			name: "TestGetTaskWithoutResults",
			ecsClient: &MockECSAPI{
				ListTasksMock: func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
					return &ecs.ListTasksOutput{
						TaskArns: []string{},
					}, nil
				},
			},
			expected: nil,
		},
	}

	for _, c := range cases {
		input := CreateMockApp(c.ecsClient)
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
		client   *MockECSAPI
		task     *ecsTypes.Task
		expected *ecsTypes.Container
	}{
		{
			name:   "TestGetContainerWithMultipleContainers",
			client: &MockECSAPI{},
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
			name:   "TestGetContainerWithSingleContainer",
			client: &MockECSAPI{},
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
		input := CreateMockApp(c.client)
		input.task = *c.task
		input.getContainer()
		if ok := assert.Equal(t, c.expected, input.container); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}
