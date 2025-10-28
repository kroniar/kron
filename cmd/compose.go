package cmd
// import "github.com/kroniar/kron/utils"
import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ComposeConfig struct {
	Name        string   `yaml:"name"`
	ComposeFiles []string `yaml:"compose_files"` // list of compose files (paths)
	Description string   `yaml:"description"`
}

// Helper: configs directory
func configsDir() string {
	return "configs"
}

func loadConfig(name string) (*ComposeConfig, error) {
	p := filepath.Join(configsDir(), name+".yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var c ComposeConfig
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func listConfigs() ([]string, error) {
	dir := configsDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var names []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasSuffix(f.Name(), ".yaml") {
			names = append(names, strings.TrimSuffix(f.Name(), ".yaml"))
		}
	}
	return names, nil
}

func ensureConfigsDir() error {
	dir := configsDir()
	return os.MkdirAll(dir, 0755)
}

func runDockerCompose(composeFiles []string, args ...string) error {
	
	if len(composeFiles) == 0 {
		return errors.New("no compose files provided")
	}
	cmdArgs := []string{}
	for _, f := range composeFiles {
		cmdArgs = append(cmdArgs, "-f", f)
	}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("docker-compose", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func dockerInspect(container string, format string) (string, error) {
	// e.g. docker inspect -f '{{.State.Health.Status}}' <container>
	args := []string{"inspect", "-f", format, container}
	cmd := exec.Command("docker", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func dockerPsFilter(filter string) (string, error) {
	// docker ps --filter label=com.kron.service=prometheus --format "{{.Names}}\t{{.Status}}"
	args := []string{"ps", "--filter", filter, "--format", "{{.Names}}\t{{.Status}}"}
	cmd := exec.Command("docker", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Manage Docker Compose stacks (kron compose <up|down|restart|ps|logs|list|add|pull|clean>)",
}

var composeUpCmd = &cobra.Command{
	Use:   "up [service]",
	Short: "Start a registered compose service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		conf, err := loadConfig(service)
		if err != nil {
			return fmt.Errorf("load config: %v", err)
		}
		fmt.Printf("⤴️  Starting %s (compose files: %v)\n", service, conf.ComposeFiles)
		return runDockerCompose(conf.ComposeFiles, "up", "-d")
	},
}

var composeDownCmd = &cobra.Command{
	Use:   "down [service]",
	Short: "Stop a registered compose service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		conf, err := loadConfig(service)
		if err != nil {
			return fmt.Errorf("load config: %v", err)
		}
		fmt.Printf("⤵️  Stopping %s\n", service)
		return runDockerCompose(conf.ComposeFiles, "down")
	},
}

var composeRestartCmd = &cobra.Command{
	Use:   "restart [service|all]",
	Short: "Restart a service or all known services",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		if target == "all" {
			names, err := listConfigs()
			if err != nil {
				return err
			}
			for _, n := range names {
				fmt.Printf("🔁 Restarting %s\n", n)
				conf, err := loadConfig(n)
				if err != nil {
					fmt.Printf("  - skip %s: %v\n", n, err)
					continue
				}
				if err := runDockerCompose(conf.ComposeFiles, "restart"); err != nil {
					fmt.Printf("  - failed %s: %v\n", n, err)
				}
			}
			return nil
		}
		conf, err := loadConfig(target)
		if err != nil {
			return fmt.Errorf("load config: %v", err)
		}
		return runDockerCompose(conf.ComposeFiles, "restart")
	},
}

var composeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available registered compose services",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := listConfigs()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Println("No services registered. Add one with 'kron compose add <name> --file <compose-file>'")
			return nil
		}
		fmt.Println("Registered services:")
		for _, n := range names {
			fmt.Printf(" - %s\n", n)
		}
		return nil
	},
}

var composePsCmd = &cobra.Command{
	Use:   "ps",
	Short: "Show containers and status for registered services",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := listConfigs()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Println("No services registered.")
			return nil
		}
		for _, n := range names {
			conf, err := loadConfig(n)
			if err != nil {
				fmt.Printf("%s: error loading config: %v\n", n, err)
				continue
			}
			fmt.Printf("== %s ==\n", n)
			// Use docker-compose ps to list containers for this compose file
			outBuf := &bytes.Buffer{}
			cmdArgs := []string{}
			for _, f := range conf.ComposeFiles {
				cmdArgs = append(cmdArgs, "-f", f)
			}
			cmdArgs = append(cmdArgs, "ps", "--services", "--filter", "status=running")
			// first list service names (docker-compose ps --services)
			c1 := exec.Command("docker-compose", cmdArgs...)
			c1.Stdout = outBuf
			c1.Stderr = os.Stderr
			if err := c1.Run(); err != nil {
				fmt.Printf("  docker-compose ps failed: %v\n", err)
				continue
			}
			services := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
			for _, s := range services {
				if strings.TrimSpace(s) == "" {
					continue
				}
				// show docker ps filter by name (container names usually compose_service_1)
				// We'll do a docker ps --filter name=<s> to find status
				out, _ := dockerPsFilter("name=" + s)
				if strings.TrimSpace(out) == "" {
					fmt.Printf("  %s  (not running or unknown)\n", s)
				} else {
					lines := strings.Split(strings.TrimSpace(out), "\n")
					for _, l := range lines {
						fmt.Printf("  %s\n", l)
					}
				}
			}
		}
		return nil
	},
}

var composeLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Stream logs for a registered service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		follow, _ := cmd.Flags().GetBool("follow")
		conf, err := loadConfig(service)
		if err != nil {
			return fmt.Errorf("load config: %v", err)
		}
		// run docker-compose logs -f service
		cmdArgs := []string{}
		for _, f := range conf.ComposeFiles {
			cmdArgs = append(cmdArgs, "-f", f)
		}
		cmdArgs = append(cmdArgs, "logs")
		if follow {
			cmdArgs = append(cmdArgs, "-f")
		}
		// stream
		c := exec.Command("docker-compose", cmdArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var composeAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Register a local compose stack config (creates configs/<name>.yaml)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		file, _ := cmd.Flags().GetString("file")
		if file == "" {
			return fmt.Errorf("--file path required")
		}
		if err := ensureConfigsDir(); err != nil {
			return err
		}
		conf := ComposeConfig{
			Name:        name,
			ComposeFiles: []string{file},
			Description: fmt.Sprintf("Registered via kron compose add %s", name),
		}
		data, _ := yaml.Marshal(&conf)
		path := filepath.Join(configsDir(), name+".yaml")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		fmt.Printf("✅ Registered %s -> %s\n", name, path)
		return nil
	},
}

var composePullCmd = &cobra.Command{
	Use:   "pull [name]",
	Short: "Download a compose file from URL and register it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		url, _ := cmd.Flags().GetString("url")
		if url == "" {
			return fmt.Errorf("--url required")
		}
		// download
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("download failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("bad response: %s", resp.Status)
		}
		if err := ensureConfigsDir(); err != nil {
			return err
		}
		// save file under examples/<name>-docker-compose.yaml
		examplesDir := "examples"
		_ = os.MkdirAll(examplesDir, 0755)
		outPath := filepath.Join(examplesDir, fmt.Sprintf("%s-docker-compose.yaml", name))
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			return err
		}
		// register
		conf := ComposeConfig{
			Name:        name,
			ComposeFiles: []string{outPath},
			Description: fmt.Sprintf("Pulled from %s", url),
		}
		data, _ := yaml.Marshal(&conf)
		path := filepath.Join(configsDir(), name+".yaml")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
		fmt.Printf("✅ Pulled and registered %s -> %s\n", name, path)
		return nil
	},
}

var composeCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Attempt to clean stopped containers and unused volumes (for registered stacks)",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := listConfigs()
		if err != nil {
			return err
		}
		for _, n := range names {
			conf, err := loadConfig(n)
			if err != nil {
				fmt.Printf("skip %s: %v\n", n, err)
				continue
			}
			// run docker-compose rm -f
			fmt.Printf("Cleaning %s\n", n)
			if err := runDockerCompose(conf.ComposeFiles, "rm", "-f"); err != nil {
				fmt.Printf("  rm failed: %v\n", err)
			}
		}
		// run docker system prune -f (optional; commented)
		// exec.Command("docker", "system", "prune", "-f").Run()
		return nil
	},
}

func init() {
	// parent
	rootCmd.AddCommand(composeCmd)

	// subcommands
	composeCmd.AddCommand(composeUpCmd)
	composeCmd.AddCommand(composeDownCmd)
	composeCmd.AddCommand(composeRestartCmd)
	composeCmd.AddCommand(composeListCmd)
	composeCmd.AddCommand(composePsCmd)
	composeCmd.AddCommand(composeLogsCmd)
	composeCmd.AddCommand(composeAddCmd)
	composeCmd.AddCommand(composePullCmd)
	composeCmd.AddCommand(composeCleanCmd)

	// flags
	composeLogsCmd.Flags().BoolP("follow", "f", false, "Follow logs")
	composeAddCmd.Flags().StringP("file", "", "", "Path to docker-compose file")
	composePullCmd.Flags().StringP("url", "u", "", "URL to raw docker-compose yaml")
}
