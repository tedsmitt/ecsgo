package app

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var (
	region   string
	endpoint string

	Red     = color.New(color.FgRed).SprintFunc()
	Magenta = color.New(color.FgMagenta).SprintFunc()
	Cyan    = color.New(color.FgCyan).SprintFunc()
	Green   = color.New(color.FgGreen).SprintFunc()
	Yellow  = color.New(color.FgYellow).SprintFunc()

	pageSize      = 15
	backOpt       = "⏎ Back" // backOpt is used to allow the user to navigate backwards in the selection prompt
	awsMaxResults = aws.Int32(int32(100))
)

func createOpts(opts []string) []string {
	initialOpts := []string{backOpt}
	return append(initialOpts, opts...)
}

func createEcsClient() *ecs.Client {
	region := viper.GetString("region")
	endpointUrl := viper.GetString("aws-endpoint-url")
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(viper.GetString("profile")),
		config.WithRegion(region),
	)
	if err != nil {
		panic(err)
	}
	client := ecs.NewFromConfig(cfg)

	return client
}

func createEc2Client() *ec2.Client {
	region := viper.GetString("region")
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(viper.GetString("profile")),
		config.WithRegion(region),
	)
	if err != nil {
		panic(err)
	}
	client := ec2.NewFromConfig(cfg)

	return client
}

func createSSMClient() *ssm.SSM {
	region := viper.GetString("region")
	endpointUrl := viper.GetString("aws-endpoint-url")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region), Endpoint: aws.String(endpointUrl)},
		Profile:           viper.GetString("profile"),
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ssm.New(sess)

	return client
}

// getPlatformFamily checks an ECS tasks properties to see if the OS can be derived from its properties, otherwise
// it will check the container instance itself to determine the OS.
func getPlatformFamily(client *ecs.Client, clusterName string, task *ecsTypes.Task) (string, error) {
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
func getContainerInstanceOS(ecsClient *ecs.Client, ec2Client *ec2.Client, cluster string, containerInstanceArn string) (string, error) {
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
	instance, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			*instanceId,
		},
	})
	operatingSystem := *instance.Reservations[0].Instances[0].PlatformDetails
	return operatingSystem, nil
}

// runCommand executes a command with args
func runCommand(process string, args ...string) error {
	if flag.Lookup("test.v") != nil {
		// emulate successful return for testing purposes
		return nil
	}

	// Capture any SIGINTs and discard them
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT)
	go func() {
		for {
			select {
			case <-sigs:
			}
		}
	}()
	defer close(sigs)

	cmd := exec.Command(process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func getContainerPort(client *ecs.Client, taskDefinitionArn string, containerName string) (*int32, error) {
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
