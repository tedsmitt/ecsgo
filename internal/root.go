package cli

import (
	"fmt"
	"os"

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
		e := CreateExecCommand()
		if err := e.Start(); err != nil {
			fmt.Printf("\n%s\n", red(err))
			os.Exit(1)
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
