package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags
var setValues []string

var yamlCmd = &cobra.Command{
	Use:   "yaml [file]",
	Short: "Modify values inside a YAML file",
	Long: `Modify one or more keys inside a YAML file.

Example:
  kron yaml config.yaml --set replicas=3 --set image.tag=v2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("please provide a YAML file path")
		}
		filePath := args[0]

		// Read YAML file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		// Parse YAML
		var content map[string]interface{}
		err = yaml.Unmarshal(data, &content)
		if err != nil {
			return fmt.Errorf("invalid YAML: %v", err)
		}

		// Apply key=value changes
		for _, kv := range setValues {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid --set value: %s", kv)
			}
			key, value := parts[0], parts[1]
			content[key] = value
		}

		// Write back to file
		newData, err := yaml.Marshal(&content)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %v", err)
		}
		err = os.WriteFile(filePath, newData, 0644)
		if err != nil {
			return fmt.Errorf("failed to save file: %v", err)
		}

		fmt.Println("✅ YAML file updated successfully!")
		return nil
	},
}

func init() {
	yamlCmd.Flags().StringArrayVar(&setValues, "set", []string{}, "key=value pairs to set in YAML")
	rootCmd.AddCommand(yamlCmd)
}
