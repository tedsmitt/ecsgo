/* aws.go contains AWS Client creation funcs and other helpers used by the main app */

package app

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/spf13/viper"
)

type EC2Client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

type ECSClient interface {
	ListClusters(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error)
	ListServices(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error)
	ListTasks(ctx context.Context, params *ecs.ListTasksInput, optFns ...func(*ecs.Options)) (*ecs.ListTasksOutput, error)
	DescribeTasks(ctx context.Context, params *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error)
	DescribeTaskDefinition(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error)
	DescribeContainerInstances(ctx context.Context, params *ecs.DescribeContainerInstancesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeContainerInstancesOutput, error)
	ExecuteCommand(ctx context.Context, params *ecs.ExecuteCommandInput, optFns ...func(*ecs.Options)) (*ecs.ExecuteCommandOutput, error)
}

func createEcsClient() *ecs.Client {
	region := viper.GetString("region")
	getCustomAWSEndpoint := func(o *ecs.Options) {
		endpointUrl := viper.GetString("aws-endpoint-url")
		if endpointUrl != "" {
			o.BaseEndpoint = aws.String(endpointUrl)
		}
	}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(viper.GetString("profile")),
		config.WithRegion(region),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxBackoffDelay(retry.NewStandard(), time.Second*1)
		}),
	)
	if err != nil {
		panic(err)
	}
	client := ecs.NewFromConfig(cfg, getCustomAWSEndpoint)

	return client
}

func createEC2Client() *ec2.Client {
	region := viper.GetString("region")
	getCustomAWSEndpoint := func(o *ec2.Options) {
		endpointUrl := viper.GetString("aws-endpoint-url")
		if endpointUrl != "" {
			o.BaseEndpoint = aws.String(endpointUrl)
		}
	}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(viper.GetString("profile")),
		config.WithRegion(region),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxBackoffDelay(retry.NewStandard(), time.Second*1)
		}),
	)
	if err != nil {
		panic(err)
	}
	client := ec2.NewFromConfig(cfg, getCustomAWSEndpoint)

	return client
}

// getPlatformFamily checks an ECS tasks properties to see if the OS can be derived from its properties, otherwise
// it will check the container instance itself to determine the OS.
func getPlatformFamily(client ECSClient, task *ecsTypes.Task) (string, error) {
	taskDefinition, err := client.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: task.TaskDefinitionArn,
	})
	if err != nil {
		return "", err
	}
	if taskDefinition.TaskDefinition.RuntimePlatform != nil {
		return string(taskDefinition.TaskDefinition.RuntimePlatform.OperatingSystemFamily), nil
	}
	return "", nil
}

// getContainerInstanceOS describes the specified container instance and checks against the backing EC2 instance
// to determine the platform.
func getContainerInstanceOS(ecsClient ECSClient, ec2Client EC2Client, cluster string, containerInstanceArn string) (string, error) {
	res, err := ecsClient.DescribeContainerInstances(context.TODO(), &ecs.DescribeContainerInstancesInput{
		Cluster: aws.String(cluster),
		ContainerInstances: []string{
			*aws.String(containerInstanceArn),
		},
	})
	if err != nil {
		return "", err
	}
	instanceId := res.ContainerInstances[0].Ec2InstanceId
	instance, _ := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			*instanceId,
		},
	})
	operatingSystem := *instance.Reservations[0].Instances[0].PlatformDetails
	return operatingSystem, nil
}

func getContainerPort(client ECSClient, taskDefinitionArn string, containerName string) (*int32, error) {
	res, err := client.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionArn),
	})
	if err != nil {
		return nil, err
	}
	var container ecsTypes.ContainerDefinition
	for _, c := range res.TaskDefinition.ContainerDefinitions {
		if *c.Name == containerName {
			container = c
		}
	}
	return container.PortMappings[0].ContainerPort, nil
}
