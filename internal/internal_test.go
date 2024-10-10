package app

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/service/ec2"
// 	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
// 	"github.com/aws/aws-sdk-go-v2/service/ecs"
// 	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
// 	"github.com/stretchr/testify/assert"
// )

// func TestGetPlatformFamily(t *testing.T) {
// 	cases := []struct {
// 		name          string
// 		expected      string
// 		mockECSClient *MockECSAPI
// 		cluster       string
// 		task          *ecsTypes.Task
// 	}{
// 		{
// 			name:    "TestGetPlatformFamilyWithFargateTask",
// 			cluster: "test",
// 			task: &ecsTypes.Task{
// 				TaskArn:        aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
// 				LaunchType:     ecsTypes.LaunchTypeFargate,
// 				PlatformFamily: aws.String("Linux"),
// 			},
// 			mockECSClient: &MockECSAPI{
// 				DescribeTaskDefinitionMock: func(ctx context.Context, input *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
// 					return &ecs.DescribeTaskDefinitionOutput{
// 						TaskDefinition: &ecsTypes.TaskDefinition{
// 							RuntimePlatform: &ecsTypes.RuntimePlatform{
// 								OperatingSystemFamily: ecsTypes.OSFamilyLinux,
// 							},
// 						},
// 					}, nil
// 				},
// 			},
// 			expected: "Linux/UNIX",
// 		},
// 		{
// 			name:    "TestGetPlatformFamilyWithEC2LaunchTaskNoRuntimePlatformFail",
// 			cluster: "test",
// 			task: &ecsTypes.Task{
// 				TaskArn:              aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
// 				LaunchType:           ecsTypes.LaunchTypeEc2,
// 				ContainerInstanceArn: aws.String("abcdefghij1234567890"),
// 			},
// 			mockECSClient: &MockECSAPI{
// 				DescribeTaskDefinitionMock: func(ctx context.Context, input *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
// 					return &ecs.DescribeTaskDefinitionOutput{
// 						TaskDefinition: &ecsTypes.TaskDefinition{},
// 					}, nil
// 				},
// 			},
// 			expected: "",
// 		},
// 	}

// 	for _, c := range cases {
// 		res, _ := getPlatformFamily(&c.mockECSClient.Client, c.cluster, c.task)
// 		if ok := assert.Equal(t, c.expected, res); ok != true {
// 			fmt.Printf("%s FAILED\n", c.name)
// 		}
// 		fmt.Printf("%s PASSED\n", c.name)
// 	}
// }

// func TestGetContainerInstanceOS(t *testing.T) {
// 	cases := []struct {
// 		name                 string
// 		expected             string
// 		mockECSClient        *MockECSAPI
// 		mockEC2Client        *MockEC2API
// 		cluster              string
// 		containerInstanceArn string
// 	}{
// 		{
// 			name:                 "TestGetContainerInstanceOS",
// 			cluster:              "test",
// 			containerInstanceArn: "abcdef123456",
// 			mockECSClient: &MockECSAPI{
// 				DescribeContainerInstancesMock: func(ctx context.Context, input *ecs.DescribeContainerInstancesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeContainerInstancesOutput, error) {
// 					return &ecs.DescribeContainerInstancesOutput{
// 						ContainerInstances: []ecsTypes.ContainerInstance{
// 							{
// 								Ec2InstanceId: aws.String("i-0063cc3b62343f4d1"),
// 							},
// 						},
// 					}, nil
// 				},
// 			},
// 			mockEC2Client: &MockEC2API{
// 				DescribeInstancesMock: func(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
// 					return &ec2.DescribeInstancesOutput{
// 						Reservations: []ec2Types.Reservation{
// 							{
// 								Instances: []ec2Types.Instance{
// 									{
// 										InstanceId:      aws.String("i-0063cc3b62343f4d1"),
// 										PlatformDetails: aws.String("Linux/UNIX"),
// 									},
// 								},
// 							},
// 						}}, nil
// 				},
// 			},
// 			expected: "Linux/UNIX",
// 		},
// 	}

// 	for _, c := range cases {
// 		res, _ := getContainerInstanceOS(&c.mockECSClient.Client, &c.mockEC2Client.Client, c.cluster, c.containerInstanceArn)
// 		if ok := assert.Equal(t, c.expected, res); ok != true {
// 			fmt.Printf("%s FAILED\n", c.name)
// 		}
// 		fmt.Printf("%s PASSED\n", c.name)
// 	}
// }
