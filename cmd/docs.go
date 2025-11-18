package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "gendocs",
	Short: "Generate markdown docs for all commands",
	Run: func(cmd *cobra.Command, args []string) {
		err := doc.GenMarkdownTree(rootCmd, "./docs/commands")
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
