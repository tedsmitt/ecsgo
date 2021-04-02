/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ecsgo",
	Short: "Tool to list and connect to your ECS tasks",
	Long: `
##########################################################
  ___  ___ ___  __ _  ___  
 / _ \/ __/ __|/ _  |/ _ \ 
|  __/ (__\__ \ (_| | (_) |
 \___|\___|___/\__, |\___/ 
               |___/       
##########################################################

Lists your ECS Clusters/tasks/containers and allows you to interactively select which to connect to. Makes use 
of the ECS ExecuteCommand API under the hood.

Requires pre-existing installation of the session-manager-plugin
(https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)
------------`,
	Version: getVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		client := createEcsClient()
		if err := StartExecuteCommand(client); err != nil {
			log.Println(red(err))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringP("task", "t", "", "Task ID to connect to")
	rootCmd.PersistentFlags().StringP("cmd", "c", "", "Command to run on the container")
	rootCmd.PersistentFlags().StringP("profile", "p", "", "AWS Profile")
	rootCmd.PersistentFlags().StringP("region", "r", "", "AWS Region")

	viper.BindPFlag("cmd", rootCmd.PersistentFlags().Lookup("cmd"))
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
}
