// Copyright © 2022 Christian R. Vozar ⚜
// Licensed under the MIT License. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/christianvozar/criprof"

	"github.com/spf13/cobra"
)

// hintsCmd represents the hints command
var hintsCmd = &cobra.Command{
	Use:   "hints",
	Short: "Display container runtime information",
	Long:  `Display container runtime information`,
	Run: func(cmd *cobra.Command, args []string) {
		i := criprof.New()

		fmt.Println(i.JSON())
	},
}

func init() {
	rootCmd.AddCommand(hintsCmd)
}
