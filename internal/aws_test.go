package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/assert"
)

type EC2ClientMock struct {
	DescribeInstancesMock func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func (m EC2ClientMock) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.DescribeInstancesMock(ctx, params, optFns...)
}

func TestGetPlatformFamily(t *testing.T) {
	cases := []struct {
		name     string
		expected string
		client   func(t *testing.T) ECSClient
		cluster  string
		task     *ecsTypes.Task
	}{
		{
			name:    "TestGetPlatformFamilyWithFargateTask",
			cluster: "test",
			task: &ecsTypes.Task{
				TaskArn:        aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				LaunchType:     ecsTypes.LaunchTypeFargate,
				PlatformFamily: aws.String("Linux"),
			},
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					DescribeTaskDefinitionMock: func(ctx context.Context, input *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
						return &ecs.DescribeTaskDefinitionOutput{
							TaskDefinition: &ecsTypes.TaskDefinition{
								RuntimePlatform: &ecsTypes.RuntimePlatform{
									OperatingSystemFamily: ecsTypes.OSFamilyLinux,
								},
							},
						}, nil
					},
				}
			},
			expected: "LINUX",
		},
		{
			name:    "TestGetPlatformFamilyWithEC2LaunchTaskNoRuntimePlatformFail",
			cluster: "test",
			task: &ecsTypes.Task{
				TaskArn:              aws.String("arn:aws:ecs:eu-west-1:111111111111:task/App/8a58117dac38436ba5547e9da5d3ac3d"),
				LaunchType:           ecsTypes.LaunchTypeEc2,
				ContainerInstanceArn: aws.String("abcdefghij1234567890"),
			},
			client: func(t *testing.T) ECSClient {
				return ECSClientMock{
					DescribeTaskDefinitionMock: func(ctx context.Context, input *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {
						return &ecs.DescribeTaskDefinitionOutput{
							TaskDefinition: &ecsTypes.TaskDefinition{},
						}, nil
					},
				}
			},
			expected: "",
		},
	}

	for _, c := range cases {
		client := c.client(t)
		res, _ := getPlatformFamily(client, c.task)
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
		ecsClient            func(t *testing.T) ECSClient
		ec2Client            func(t *testing.T) EC2Client
		cluster              string
		containerInstanceArn string
	}{
		{
			name:                 "TestGetContainerInstanceOS",
			cluster:              "test",
			containerInstanceArn: "abcdef123456",
			ecsClient: func(t *testing.T) ECSClient {
				return ECSClientMock{
					DescribeContainerInstancesMock: func(ctx context.Context, input *ecs.DescribeContainerInstancesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeContainerInstancesOutput, error) {
						return &ecs.DescribeContainerInstancesOutput{
							ContainerInstances: []ecsTypes.ContainerInstance{
								{
									Ec2InstanceId: aws.String("i-0063cc3b62343f4d1"),
								},
							},
						}, nil
					},
				}
			},
			ec2Client: func(t *testing.T) EC2Client {
				return EC2ClientMock{
					DescribeInstancesMock: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
						return &ec2.DescribeInstancesOutput{
							Reservations: []ec2Types.Reservation{
								{
									Instances: []ec2Types.Instance{
										{
											InstanceId:      aws.String("i-0063cc3b62343f4d1"),
											PlatformDetails: aws.String("Linux/UNIX"),
										},
									},
								},
							}}, nil
					},
				}
			},
			expected: "Linux/UNIX",
		},
	}

	for _, c := range cases {
		ecsClient := c.ecsClient(t)
		ec2Client := c.ec2Client(t)
		res, _ := getContainerInstanceOS(ecsClient, ec2Client, c.cluster, c.containerInstanceArn)
		if ok := assert.Equal(t, c.expected, res); ok != true {
			fmt.Printf("%s FAILED\n", c.name)
		}
		fmt.Printf("%s PASSED\n", c.name)
	}
}
