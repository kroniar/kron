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



type SetupTool struct {
	Description string            `yaml:"description"`
	InstallCmd  map[string]string `yaml:"install_cmd"` // OS-specific install commands
}
type SetupConfig struct {
	Tools map[string]SetupTool `yaml:"tools"`
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

func init() {
	setupCmd.Flags().BoolVarP(&forceReinstall, "force", "f", false, "Force reinstall even if already installed")
	rootCmd.AddCommand(setupCmd)
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

	// Auto-create default config if missing
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(configPath), 0755)
		createDefaultSetupConfig(configPath)
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

func createDefaultSetupConfig(path string) {
	defaultConfig := SetupConfig{
		Tools: map[string]SetupTool{
			"docker": {
				Description: "Install Docker Engine",
				InstallCmd: map[string]string{
					"linux":   "sudo apt update && sudo apt install -y ca-certificates curl gnupg lsb-release && sudo install -m 0755 -d /etc/apt/keyrings && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg && echo \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null && sudo apt update -y && sudo apt install -y docker-ce docker-ce-cli containerd.io && sudo systemctl enable docker && sudo systemctl start docker",
					"darwin":  "brew install docker",
					"windows": "choco install docker-desktop",
				},
			},
			"docker-compose": {
				Description: "Install Docker Compose CLI plugin",
				InstallCmd: map[string]string{
					"linux":   "DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker} && mkdir -p $DOCKER_CONFIG/cli-plugins && LATEST=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep browser_download_url | grep linux-x86_64 | cut -d '\"' -f 4) && curl -SL $LATEST -o $DOCKER_CONFIG/cli-plugins/docker-compose && chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose && docker compose version",
					"darwin":  "brew install docker-compose",
					"windows": "choco install docker-compose",
				},
			},
			"kubectl": {
				Description: "Install Kubernetes CLI",
				InstallCmd: map[string]string{
					"linux":   "sudo snap install kubectl --classic",
					"darwin":  "brew install kubectl",
					"windows": "choco install kubernetes-cli",
				},
			},
			"vscode": {
				Description: "Install Visual Studio Code",
				InstallCmd: map[string]string{
					"linux":   "sudo snap install code --classic",
					"darwin":  "brew install --cask visual-studio-code",
					"windows": "choco install vscode",
				},
			},
			"prometheus": {
				Description: "Install Prometheus monitoring system",
				InstallCmd: map[string]string{
					"linux":   "sudo apt install -y prometheus",
					"darwin":  "brew install prometheus",
					"windows": "choco install prometheus",
				},
			},
			"grafana": {
				Description: "Install Grafana dashboard",
				InstallCmd: map[string]string{
					"linux":   "sudo apt install -y grafana",
					"darwin":  "brew install grafana",
					"windows": "choco install grafana",
				},
			},
		},
	}

	data, _ := yaml.Marshal(&defaultConfig)
	_ = os.WriteFile(path, data, 0644)
	fmt.Printf("🧩 Created default setup config at %s\n", path)
}
