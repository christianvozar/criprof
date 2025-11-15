// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.

// Package cmd provides the command-line interface for criprof.
//
// This package implements the CLI commands using the Cobra framework, allowing
// users to interact with criprof from the command line to detect and report
// container runtime information.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "criprof",
	Short: "Container Runtime Interface profiling and introspection.",
	Long:  `Container Runtime Interface profiling and introspection.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute runs the root command and handles any errors that occur.
//
// This is the main entry point for the CLI application, called by main.main().
// It executes the root command and all registered subcommands, processing
// command-line flags and arguments.
//
// If an error occurs during command execution, it is printed to stdout and
// the program exits with status code 1.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.criprof.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads configuration from file and environment variables.
//
// This function is called during Cobra initialization. It attempts to load
// configuration from:
//   - The config file specified by --config flag
//   - $HOME/.criprof.yaml (default location)
//   - Environment variables that match configuration keys
//
// If a config file is found and successfully loaded, its path is printed to stdout.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".criprof" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".criprof")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
