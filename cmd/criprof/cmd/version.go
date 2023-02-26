// Copyright © 2022-2023 Christian R. Vozar
// Licensed under the MIT License. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("criprof version 1.1")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
