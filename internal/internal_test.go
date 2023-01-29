package app

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
)

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
