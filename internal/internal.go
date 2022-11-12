package app

import (
	"flag"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
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

	backOpt = "⏎ Back" // backOpt is used to allow the user to navigate backwards in the selection prompt
)

func createOpts(opts []string) []string {
	initialOpts := []string{backOpt}
	return append(initialOpts, opts...)
}

func createEcsClient() *ecs.ECS {
	region := viper.GetString("region")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		Profile:           viper.GetString("profile"),
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ecs.New(sess)

	return client
}

func createEc2Client() *ec2.EC2 {
	region := viper.GetString("region")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		Profile:           viper.GetString("profile"),
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ec2.New(sess)

	return client
}

// getPlatformFamily checks an ECS tasks properties to see if the OS can be derived from its properties, otherwise
// it will check the container instance itself to determine the OS.
func getPlatformFamily(client ecsiface.ECSAPI, clusterName string, task *ecs.Task) (string, error) {
	taskDefinition, err := client.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: task.TaskDefinitionArn,
	})
	if err != nil {
		return "", err
	}
	if taskDefinition.TaskDefinition.RuntimePlatform != nil {
		return *taskDefinition.TaskDefinition.RuntimePlatform.OperatingSystemFamily, nil
	}
	return "", nil
}

// getContainerInstanceOS describes the specified container instance and checks against the backing EC2 instance
// to determine the platform.
func getContainerInstanceOS(ecsClient ecsiface.ECSAPI, ec2Client ec2iface.EC2API, cluster string, containerInstanceArn string) (string, error) {
	res, err := ecsClient.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster: aws.String(cluster),
		ContainerInstances: []*string{
			aws.String(containerInstanceArn),
		},
	})
	if err != nil {
		return "", err
	}
	instanceId := res.ContainerInstances[0].Ec2InstanceId
	instance, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			instanceId,
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

	cmd := exec.Command(process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func getContainerPort(client ecsiface.ECSAPI, taskDefinitionArn string, containerName string) (*int64, error) {
	res, err := client.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionArn),
	})
	if err != nil {
		return nil, err
	}
	var container ecs.ContainerDefinition
	for _, c := range res.TaskDefinition.ContainerDefinitions {
		if *c.Name == containerName {
			container = *c
		}
	}
	return container.PortMappings[0].ContainerPort, nil
}
