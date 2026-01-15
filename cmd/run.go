/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/kroniar/kron/src/run"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:                "run",
	Short:              "A powerfull command to run various scripts and modules",
	Long:               `Modules description willbe added here in the future.`,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		run.RunInit(args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
