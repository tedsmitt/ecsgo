package app

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("AWS_DEFAULT_REGION", "eu-west-1")
}

type MockECSAPI struct {
	ecsiface.ECSAPI                // embedding of the interface is needed to skip implementation of all methods
	ListClustersMock               func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
	ListServicesMock               func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error)
	ListTasksMock                  func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error)
	DescribeTasksMock              func(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error)
	DescribeTaskDefinitionMock     func(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error)
	DescribeContainerInstancesMock func(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
	ExecuteCommandMock             func(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error)
}

func (m *MockECSAPI) ListClusters(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	if m.ListClustersMock != nil {
		return m.ListClustersMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) ListServices(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
	if m.ListServicesMock != nil {
		return m.ListServicesMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	if m.ListTasksMock != nil {
		return m.ListTasksMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	if m.DescribeTasksMock != nil {
		return m.DescribeTasksMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) DescribeTaskDefinition(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
	if m.DescribeTaskDefinitionMock != nil {
		return m.DescribeTaskDefinitionMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	if m.DescribeContainerInstancesMock != nil {
		return m.DescribeContainerInstancesMock(input)
	}
	return nil, nil
}

func (m *MockECSAPI) ExecuteCommand(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) {
	if m.ExecuteCommandMock != nil {
		return m.ExecuteCommandMock(input)
	}
	return nil, nil
}

type MockEC2API struct {
	ec2iface.EC2API       // embedding of the interface is needed to skip implementation of all methods
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
		client:   c,
		region:   "eu-west-1",
		endpoint: "ecs.eu-west-1.amazonaws.com",
	}

	return e
}

func TestGetCluster(t *testing.T) {
	paginationCall := 0
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
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/App"),
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/blueGreen-1"),
						},
					}, nil
				},
			},
			expected: "App",
		},
		{
			name: "TestGetClusterWithResultsPaginated",
			client: &MockECSAPI{
				ListClustersMock: func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
					var clusters []*string
					for i := paginationCall; i < (paginationCall * 100); i++ {
						clusters = append(clusters, aws.String(fmt.Sprintf("arn:aws:ecs:eu-west-1:1111111111:cluster/test-cluster-%d", i)))
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
		{
			name: "TestGetClusterWithSingleResult",
			client: &MockECSAPI{
				ListClustersMock: func(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
					return &ecs.ListClustersOutput{
						ClusterArns: []*string{
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/App"),
						},
					}, nil
				},
			},
			expected: "App",
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
		input := CreateMockApp(c.client)
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
		client   *MockECSAPI
		expected string
	}{
		{
			name: "TestGetServiceWithResults",
			client: &MockECSAPI{
				ListServicesMock: func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
					return &ecs.ListServicesOutput{
						ServiceArns: []*string{
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/App/test-service-1"),
							aws.String("arn:aws:ecs:eu-west-1:1111111111:cluster/blueGreen/test-service-2"),
						},
					}, nil
				},
			},
			expected: "test-service-1",
		},
		{
			name: "TestGetServiceWithResultsPaginated",
			client: &MockECSAPI{
				ListServicesMock: func(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
					var services []*string
					for i := paginationCall; i < (paginationCall * 100); i++ {
						services = append(services, aws.String(fmt.Sprintf("arn:aws:ecs:eu-west-1:1111111111:cluster/App/test-service-%d", i)))
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
		input := CreateMockApp(c.client)
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
		expected *ecs.Task
	}{
		{
			name: "TestGetTaskWithResults",
			client: &MockECSAPI{
				ListTasksMock: func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
					return &ecs.ListTasksOutput{
						TaskArns: []*string{
							aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
						},
					}, nil
				},
				DescribeTasksMock: func(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
					var tasks []*ecs.Task
					for _, taskArn := range input.Tasks {
						tasks = append(tasks, &ecs.Task{TaskArn: taskArn, LaunchType: aws.String("FARGATE")})
					}
					return &ecs.DescribeTasksOutput{
						Tasks: tasks,
					}, nil
				},
			},
			expected: &ecs.Task{
				TaskArn:    aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				LaunchType: aws.String("FARGATE"),
			},
		},
		{
			name: "TestGetTaskWithResultsPaginated",
			client: &MockECSAPI{
				ListTasksMock: func(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
					var taskArns []*string
					var tasks []*ecs.Task
					for i := paginationCall; i < (paginationCall * 100); i++ {
						taskArn := aws.String(fmt.Sprintf("arn:aws:ecs:eu-west-1:111111111111:task/App/%d", i))
						taskArns = append(taskArns, taskArn)
						tasks = append(tasks, &ecs.Task{TaskArn: taskArn, LaunchType: aws.String("FARGATE")})
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
					var tasks []*ecs.Task
					for _, taskArn := range input.Tasks {
						tasks = append(tasks, &ecs.Task{TaskArn: taskArn, LaunchType: aws.String("FARGATE")})
					}
					return &ecs.DescribeTasksOutput{
						Tasks: tasks,
					}, nil
				},
			},
			expected: &ecs.Task{
				TaskArn:    aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/199"),
				LaunchType: aws.String("FARGATE"),
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
		input := CreateMockApp(c.client)
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
		input := CreateMockApp(c.client)
		input.task = c.task
		input.getContainer()
		if ok := assert.Equal(t, c.expected, input.container); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}

func TestExecuteInput(t *testing.T) {
	cases := []struct {
		name     string
		expected error
		client   *MockECSAPI
		cluster  string
		task     *ecs.Task
	}{
		{
			name:    "TestExecuteInput",
			cluster: "test",
			task: &ecs.Task{
				TaskArn: aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				Containers: []*ecs.Container{
					{
						Name:      aws.String("nginx"),
						RuntimeId: aws.String("544e08d919364be9926186b086c29868-2531612879"),
					},
				},
				PlatformFamily: aws.String("Linux"),
			},
			client: &MockECSAPI{
				ExecuteCommandMock: func(input *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) {
					return &ecs.ExecuteCommandOutput{
						Session: &ecs.Session{
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
			client:   c.client,
			region:   "eu-west-1",
			endpoint: "ecs.eu-west-1.amazonaws.com",
			cluster:  c.cluster,
			task:     c.task,
		}
		app.container = c.task.Containers[0]
		err := app.executeCommand()
		if ok := assert.Equal(t, c.expected, err); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}

func TestGetPlatformFamily(t *testing.T) {
	cases := []struct {
		name      string
		expected  string
		ecsClient *MockECSAPI
		cluster   string
		task      *ecs.Task
	}{
		{
			name:    "TestGetPlatformFamilyWithFargateTask",
			cluster: "test",
			task: &ecs.Task{
				TaskArn:        aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				LaunchType:     aws.String("FARGATE"),
				PlatformFamily: aws.String("Linux"),
			},
			ecsClient: &MockECSAPI{
				DescribeTaskDefinitionMock: func(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
					return &ecs.DescribeTaskDefinitionOutput{
						TaskDefinition: &ecs.TaskDefinition{
							RuntimePlatform: &ecs.RuntimePlatform{
								OperatingSystemFamily: aws.String("Linux/UNIX"),
							},
						},
					}, nil
				},
			},
			expected: "Linux/UNIX",
		},
		{
			name:    "TestGetPlatformFamilyWithEC2LaunchTaskNoRuntimePlatformFail",
			cluster: "test",
			task: &ecs.Task{
				TaskArn:              aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				LaunchType:           aws.String("EC2"),
				ContainerInstanceArn: aws.String("abcdefghij1234567890"),
			},
			ecsClient: &MockECSAPI{
				DescribeTaskDefinitionMock: func(input *ecs.DescribeTaskDefinitionInput) (*ecs.DescribeTaskDefinitionOutput, error) {
					return &ecs.DescribeTaskDefinitionOutput{
						TaskDefinition: &ecs.TaskDefinition{},
					}, nil
				},
			},
			expected: "",
		},
	}

	for _, c := range cases {
		res, _ := getPlatformFamily(c.ecsClient, c.cluster, c.task)
		if ok := assert.Equal(t, c.expected, res); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}

func TestGetContainerInstanceOS(t *testing.T) {
	cases := []struct {
		name                 string
		expected             string
		ecsClient            *MockECSAPI
		ec2Client            *MockEC2API
		cluster              string
		containerInstanceArn string
	}{
		{
			name:                 "TestGetContainerInstanceOS",
			cluster:              "test",
			containerInstanceArn: "abcdef123456",
			ecsClient: &MockECSAPI{
				DescribeContainerInstancesMock: func(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
					return &ecs.DescribeContainerInstancesOutput{
						ContainerInstances: []*ecs.ContainerInstance{
							{
								Ec2InstanceId: aws.String("i-0063cc3b62343f4d1"),
							},
						},
					}, nil
				},
			},
			ec2Client: &MockEC2API{
				DescribeInstancesMock: func(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
					return &ec2.DescribeInstancesOutput{
						Reservations: []*ec2.Reservation{
							{
								Instances: []*ec2.Instance{
									{
										InstanceId:      aws.String("i-0063cc3b62343f4d1"),
										PlatformDetails: aws.String("Linux/UNIX"),
									},
								},
							},
						}}, nil
				},
			},
			expected: "Linux/UNIX",
		},
	}

	for _, c := range cases {
		res, _ := getContainerInstanceOS(c.ecsClient, c.ec2Client, c.cluster, c.containerInstanceArn)
		if ok := assert.Equal(t, c.expected, res); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}
