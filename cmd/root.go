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
	// Execute() prints the error once, so Cobra shouldn't print it too
	SilenceErrors: true,
	// Runs after flags and args are parsed: flag/arg mistakes still show usage,
	// errors while running (bad file, failed --set, ...) don't
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	},
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
