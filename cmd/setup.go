package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ---------------------------
// 🧱 Data Structures

// ---------------------------

type SetupConfig struct {
	Tools map[string]SetupTool `yaml:"tools"`
}
type SetupTool struct {
	Description  string            `yaml:"description"`
	InstallCmd   map[string]string `yaml:"install_cmd"` // OS-specific install commands
	UninstallCmd map[string]string `yaml:"uninstall_cmd,omitempty"`
}

// ---------------------------
// ⚙️ Cobra Command
// ---------------------------

var forceReinstall bool

var setupCmd = &cobra.Command{
	Use:   "setup [tool|list]",
	Short: "Install or manage system dependencies (preloaded with popular tools)",
	Long: `Kron Setup lets you install essential DevOps tools easily.

Examples:
  kron setup list
  kron setup docker
  kron setup prometheus
  kron setup docker --force    # reinstall or update`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a tool name or use 'list'.")
			return
		}

		switch args[0] {
		case "list":
			listAllTools()
		default:
			installTool(args[0], forceReinstall)
		}
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [tool]",
	Short: "Uninstall or remove a tool installed via Kron Setup",
	Long: `Use this command to uninstall a tool defined in your setup.yaml.

Example:
  kron setup remove docker
  kron setup remove prometheus`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a tool to remove.")
			return
		}
		uninstallTool(args[0])
	},
}

func init() {
	setupCmd.Flags().BoolVarP(&forceReinstall, "force", "f", false, "Force reinstall even if already installed")
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(removeCmd)
}

// ---------------------------
// 🔧 Core Functions
// ---------------------------

func listAllTools() {
	cfg := loadSetupConfig()
	if len(cfg.Tools) == 0 {
		fmt.Println("No tools found. Your setup file might be empty.")
		return
	}

	fmt.Println("📦 Available tools:")
	for name, tool := range cfg.Tools {
		fmt.Printf("  - %s: %s\n", name, tool.Description)
	}
	fmt.Println("\nUse 'kron setup <tool>' to install or update.")
}

func installTool(toolName string, force bool) {
	cfg := loadSetupConfig()
	tool, ok := cfg.Tools[toolName]
	if !ok {
		fmt.Printf("❌ Unknown tool: %s\n", toolName)
		fmt.Println("Run 'kron setup list' to see available options.")
		return
	}
	osType := runtime.GOOS
	fmt.Println("🔧 Setting up tool:", toolName)
	fmt.Println("🔍 Detected OS:", osType)
	cmd := tool.InstallCmd[osType]
	if cmd == "" {
		fmt.Printf("❌ No install command for %s on %s.\n", toolName, osType)
		return
	}

	if checkCommandExists(toolName) && !force {
		fmt.Printf("✅ %s is already installed. Use '--force' to reinstall/update.\n", toolName)
		return
	}

	fmt.Printf("🚀 Installing or updating %s...\n", toolName)
	runCommand(cmd)
	fmt.Printf("✅ %s installation/update completed!\n", toolName)
}

func loadSetupConfig() SetupConfig {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".kron", "setup.yaml")

	cfg := SetupConfig{
		Tools: make(map[string]SetupTool),
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("❌ Failed to read setup config:", err)
		return cfg
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Println("❌ Failed to parse setup config:", err)
		return cfg
	}

	if cfg.Tools == nil {
		cfg.Tools = make(map[string]SetupTool)
	}

	return cfg
}

func uninstallTool(toolName string) {
	cfg := loadSetupConfig()
	tool, ok := cfg.Tools[toolName]
	if !ok {
		fmt.Printf("❌ Unknown tool: %s\n", toolName)
		fmt.Println("Run 'kron setup list' to see available tools.")
		return
	}

	osType := runtime.GOOS
	cmd := tool.UninstallCmd[osType]
	if cmd == "" {
		fmt.Printf("❌ No uninstall command defined for %s on %s.\n", toolName, osType)
		return
	}

	fmt.Printf("🧹 Uninstalling %s...\n", toolName)
	runCommand(cmd)
	fmt.Printf("✅ %s has been uninstalled successfully!\n", toolName)
}

// ---------------------------
// 🧰 Helpers
// ---------------------------

func checkCommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func runCommand(command string) {
	c := exec.Command("bash", "-c", command)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fmt.Printf("❌ Error running command: %v\n", err)
	}
}
