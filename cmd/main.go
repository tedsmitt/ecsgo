/*
Copyright © 2021 Ed Smith ed@edintheclouds.io

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
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	app "github.com/tedsmitt/ecsgo/internal"
)

func main() {
	rootCmd.Execute()
}

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
		a := app.CreateApp()
		if err := a.Start(); err != nil {
			fmt.Printf("\n%s\n", app.Red(err))
			os.Exit(1)
		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringP("cmd", "c", "", "Command to run on the container")
	rootCmd.PersistentFlags().StringP("profile", "p", "", "AWS Profile")
	rootCmd.PersistentFlags().StringP("region", "r", "", "AWS Region")
	rootCmd.PersistentFlags().BoolP("forward", "f", false, "Port Forward")
	rootCmd.PersistentFlags().StringP("local-port", "l", "", "Local port for use with port forwarding")

	viper.BindPFlag("cmd", rootCmd.PersistentFlags().Lookup("cmd"))
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("forward", rootCmd.PersistentFlags().Lookup("forward"))
	viper.BindPFlag("local-port", rootCmd.PersistentFlags().Lookup("local-port"))
}
