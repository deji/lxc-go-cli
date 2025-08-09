/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/lxc-go-cli/internal/logger"
)

var (
	logLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lxc-go-cli",
	Short: "lxc cli tool to create and manage containers running docker",
	Long: `lxc-go-cli is a cli tool to create and manage containers for docker.
	It is a wrapper around the lxc cli tool to create and manage containers with the
	btrfs storage backend. Docker and docker-compose are installed in the container too.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logging level from flag
		logger.SetLevelFromString(logLevel)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Add persistent log level flag
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Set the logging level (debug, info, warn, error)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
