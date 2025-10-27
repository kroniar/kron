package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kron",
	Short: "Kron — automate DevOps workflows, configs, and local setups",
	Long: `Kron is an open source DevOps automation CLI.

You can use it to:
  • Modify JSON/YAML files across projects
  • Automate GitHub/GitLab pull requests
  • Set up local Grafana/Prometheus instances
  • Expose and collect metrics easily

Examples:
  kron yaml update config.yaml --set replicas=3
  kron pr create --repo github.com/user/repo --branch feature-x
  kron setup prometheus grafana
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Kron ⚙️ — run 'kron help' to see available commands")
	},
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
