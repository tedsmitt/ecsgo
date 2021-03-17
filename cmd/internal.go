package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"

	homedir "github.com/mitchellh/go-homedir"
)

// ReadAwsConfig reads in the aws config file and returns a slice of all profile names
func ReadAwsConfig() ([]string, error) {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(fmt.Sprintf("%s/.aws/config", home))
	if err != nil {
		log.Fatal(err)
	}

	var profiles []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Index(line, "[profile ") > -1 {
			raw := strings.Split(line, " ")[1]
			profile := strings.TrimRight(raw, "]")
			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

func AskProfile(profiles []string) (string, error) {
	var profile string
	var qs = []*survey.Question{
		{
			Name: "profile",
			Prompt: &survey.Select{
				Message: "Select your AWS Profile",
				Options: profiles,
			},
		},
	}
	// perform the questions
	err := survey.Ask(qs, &profile)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return profile, nil
}

func createEcsClient() *ecs.ECS {
	mySession, _ := session.NewSessionWithOptions(session.Options{
		Profile: "edintheclouds-dev",
		Config: aws.Config{
			Region: aws.String("eu-west-1"),
		},
	})

	// Test credentials
	svc := ecs.New(mySession)

	return svc
}
