package cmd

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Flags
var setValues []string
var forceOverwrite bool

var yamlCmd = &cobra.Command{
	Use:   "yaml [file]",
	Short: "Get/Modify values inside a YAML file",
	Long: `Get/Modify one or more keys inside a YAML file.

Example:
  kron yaml config.yaml --set replicas=3 --set image.tag=v2
  kron yaml config.yaml --set image=nginx --force    # allow changing a field's type`,
	RunE: func(cmd *cobra.Command, args []string) error {
		action := "GET"

		if cmd.Flags().Changed("set") {
			action = "SET"
		}

		if len(args) <= 1 {
			if len(args) == 0 {
				return fmt.Errorf("Wrong Usage!!")
			}
			if len(args) == 1 && len(setValues) == 0 {
				return fmt.Errorf("Add flags or Arguments to GET/SET Key Values")
			}
		}
		filePath := args[0]

		// Read YAML file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		// Parse YAML
		var content yaml.Node
		err = yaml.Unmarshal(data, &content)
		if err != nil {
			return fmt.Errorf("invalid YAML: %v", err)
		}

		// An empty or comment-only file has no document at all, and "---", "~" or
		// "null" gives a null root: treat both as an empty mapping so keys can be added
		noDocument := len(content.Content) == 0
		if noDocument {
			content = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
		}
		contentRootNode := content.Content[0]
		if contentRootNode.Kind == yaml.ScalarNode && contentRootNode.Tag == "!!null" {
			*contentRootNode = yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		}

		if action == "GET" {
			keys := args[1:]
			for _, key := range keys {
				oldKeyValue, found, _ := GetNestedValue(contentRootNode, key)
				if found {
					fmt.Printf("Key-Values Pairs - \n\n%v : %v\n", key, oldKeyValue)
				} else {
					fmt.Printf("Key - `%v` does not exist in file %v\n", key, filePath)
				}
			}
		} else {
			if len(args) > 1 {
				return fmt.Errorf("unexpected argument %q: repeat --set for each key, e.g. --set a=1 --set b=2", args[1])
			}
			if contentRootNode.Kind != yaml.MappingNode {
				kind := "a single value"
				if contentRootNode.Kind == yaml.SequenceNode {
					kind = "a list"
				}
				return fmt.Errorf("the top level of %s is %s, not key: value pairs, so --set can't add keys to it", filePath, kind)
			}
			// Parse every --set first, so a typo fails before anything is applied
			type change struct{ key, value string }
			var changes []change
			for _, keyval := range setValues {
				parts := strings.SplitN(keyval, "=", 2)
				if len(parts) != 2 || parts[0] == "" {
					return fmt.Errorf("incorrect argument %q, should be like `key=value`", keyval)
				}
				changes = append(changes, change{key: parts[0], value: parts[1]})
			}

			// Remember each key's original value (once per key, in first-seen order)
			var keys []string
			oldValues := map[string]string{}
			existed := map[string]bool{}
			for _, c := range changes {
				if _, seen := existed[c.key]; seen {
					continue
				}
				keys = append(keys, c.key)
				node := findNode(contentRootNode, c.key)
				existed[c.key] = node != nil
				if node != nil {
					oldValues[c.key] = renderValue(node)
				}
			}

			// Apply all changes in memory; any failure returns before the file is written
			for _, c := range changes {
				if err := SetNestedValue(contentRootNode, c.key, c.value, forceOverwrite); err != nil {
					return fmt.Errorf("%v (no changes were saved)", err)
				}
			}

			// Compare each key's final value with its original one
			var added, updated, unchanged []string
			for _, key := range keys {
				newValue := "<removed>"
				if node := findNode(contentRootNode, key); node != nil {
					newValue = renderValue(node)
				}
				switch {
				case !existed[key]:
					added = append(added, fmt.Sprintf("  + %s: %s", key, newValue))
				case oldValues[key] != newValue:
					updated = append(updated, fmt.Sprintf("  ~ %s: changed from %s to %s", key, oldValues[key], newValue))
				default:
					unchanged = append(unchanged, fmt.Sprintf("  = %s: %s", key, newValue))
				}
			}

			if len(added) == 0 && len(updated) == 0 {
				fmt.Println("No changes: every key already has the requested value")
				printChangeLog("Unchanged", unchanged)
				return nil
			}

			// Write the whole document back with the file's own indent width
			var buf bytes.Buffer
			enc := yaml.NewEncoder(&buf)
			enc.SetIndent(detectIndent(data))
			if err := enc.Encode(&content); err != nil {
				return fmt.Errorf("failed to marshal YAML: %v", err)
			}
			enc.Close()
			newData := buf.Bytes()
			// yaml.v3 drops the comments of a comment-only file, so keep its original text on top
			if noDocument && strings.TrimSpace(string(data)) != "" {
				newData = append([]byte(strings.TrimRight(string(data), " \t\r\n")+"\n"), newData...)
			}
			if err := os.WriteFile(filePath, newData, 0644); err != nil {
				return fmt.Errorf("failed to save file: %v", err)
			}

			fmt.Println("✅ YAML file updated successfully!")
			printChangeLog("Added", added)
			printChangeLog("Updated", updated)
			printChangeLog("Unchanged", unchanged)
		}

		return nil
	},
}

func init() {
	yamlCmd.Flags().StringArrayVar(&setValues, "set", []string{}, "key=value pairs to set in YAML")
	yamlCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "allow --set to change a field's type (scalar <-> mapping/list)")
	rootCmd.AddCommand(yamlCmd)
}

// findNode returns the node at dotPath, or nil if it doesn't exist
func findNode(mapping *yaml.Node, dotPath string) *yaml.Node {
	curr := mapping
	for _, key := range strings.Split(dotPath, ".") {
		curr = findKeyInMapping(curr, key)
		if curr == nil {
			return nil
		}
	}
	return curr
}

// renderValue renders a node on one line for the change log, e.g. 8080, "8080" or {name: web}
func renderValue(n *yaml.Node) string {
	c := withoutComments(n)
	if c.Kind != yaml.ScalarNode {
		c.Style = yaml.FlowStyle
	}
	out, err := yaml.Marshal(c)
	if err != nil {
		return n.Value
	}
	return strings.TrimSpace(string(out))
}

// withoutComments returns a deep copy of n with all comments removed
func withoutComments(n *yaml.Node) *yaml.Node {
	c := *n
	c.HeadComment, c.LineComment, c.FootComment = "", "", ""
	c.Content = make([]*yaml.Node, len(n.Content))
	for i, child := range n.Content {
		c.Content[i] = withoutComments(child)
	}
	return &c
}

// printChangeLog prints one section of the --set change log, skipping empty sections
func printChangeLog(title string, lines []string) {
	if len(lines) == 0 {
		return
	}
	fmt.Printf("\n%s (%d):\n%s\n", title, len(lines), strings.Join(lines, "\n"))
}

// detectIndent returns the indent width of the first indented line (default 2)
func detectIndent(data []byte) int {
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimLeft(line, " ")
		if n := len(line) - len(trimmed); n > 0 && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return n
		}
	}
	return 2
}

// inferTagAndValue inspects a string and returns the proper YAML tag
func inferTagAndValue(val string) (string, string) {

	// Booleans
	if val == "true" || val == "false" {
		return "!!bool", val
	}

	// Integers (e.g., 8080, -42)
	if matched, _ := regexp.MatchString(`^-?\d+$`, val); matched {
		return "!!int", val
	}

	// Floats (e.g., 3.14, -0.01)
	if matched, _ := regexp.MatchString(`^-?\d+\.\d+$`, val); matched {
		return "!!float", val
	}

	// Default to String
	return "!!str", val
}

// ---------------------------------------------------------
// Helper: Traverse dot-notation path to GET a value node
// ---------------------------------------------------------
func GetNestedValue(mapping *yaml.Node, dotPath string) (string, bool, yaml.Kind) {
	keys := strings.Split(dotPath, ".")
	curr := mapping

	for idx, key := range keys {
		valNode := findKeyInMapping(curr, key)
		if valNode == nil {
			return "", false, 0
		}

		// If this is the last key, return its scalar value
		if idx == len(keys)-1 {
			var found = false
			var value = ""
			var yamlKind yaml.Kind = 0
			switch valNode.Kind {
			case yaml.ScalarNode:
				{
					value = valNode.Value
					found = true
					yamlKind = yaml.ScalarNode
				}
			case yaml.MappingNode:
				{
					value = "\n\t" + strings.Join(printMapOrder(valNode, ""), "\t")
					found = true
					yamlKind = yaml.MappingNode
				}
			case yaml.SequenceNode:
				{
					value = "\n\t" + strings.Join(printMapOrder(valNode, ""), "\t")
					found = true
					yamlKind = yaml.SequenceNode
				}
			}
			return value, found, yamlKind
		}

		// Otherwise, descend into the next mapping node
		curr = valNode
	}
	return "", false, 0
}

// SetNestedValue sets dotPath to rawValue. Without force it refuses to change
// an existing field's type (scalar <-> mapping/list); with force it overwrites.
func SetNestedValue(mapping *yaml.Node, dotPath string, rawValue string, force bool) error {
	keys := strings.Split(dotPath, ".")
	curr := mapping

	for idx, key := range keys {
		isLast := idx == len(keys)-1

		if isLast {
			tag, val := inferTagAndValue(rawValue)

			// Strings are always written double-quoted; everything else plain.
			// Must be set explicitly, otherwise an existing node keeps its old style.
			var style yaml.Style = 0
			if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
				style = yaml.DoubleQuotedStyle
				val = val[1 : len(val)-1]
			}
			if len(val) >= 2 && val[0] == '\'' && val[len(val)-1] == '\'' {
				style = yaml.SingleQuotedStyle
				val = val[1 : len(val)-1]
			}

			valNode := findKeyInMapping(curr, key)
			if valNode != nil {
				if valNode.Kind == yaml.MappingNode || valNode.Kind == yaml.SequenceNode {
					kind := "mapping"
					if valNode.Kind == yaml.SequenceNode {
						kind = "list"
					}
					if !force {
						return fmt.Errorf("%q is a %s, cannot replace it with a value (use --force to overwrite)", dotPath, kind)
					}
					fmt.Printf("⚠️  --force: replacing %s %q with a value\n", kind, dotPath)
				}
				// Update existing node
				valNode.Kind = yaml.ScalarNode
				valNode.Tag = tag
				valNode.Value = val
				valNode.Style = style
				valNode.Content = nil
			} else {
				// Create new node
				keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
				valNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: val, Style: style}
				// valNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: val, Style : 0}
				curr.Content = append(curr.Content, keyNode, valNode)
			}
			return nil
		}

		// Traverse intermediate mapping nodes: create missing ones,
		// but only convert an existing scalar/list into a mapping with force
		valNode := findKeyInMapping(curr, key)
		if valNode == nil {
			newMapNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
			curr.Content = append(curr.Content, keyNode, newMapNode)
			valNode = newMapNode
		} else if valNode.Kind != yaml.MappingNode {
			path := strings.Join(keys[:idx+1], ".")
			subKey := keys[idx+1]
			if !force {
				switch valNode.Kind {
				case yaml.ScalarNode:
					return fmt.Errorf("%q is a scalar value (%v), cannot add sub-field %q to it (use --force to overwrite)", path, valNode.Value, subKey)
				case yaml.SequenceNode:
					return fmt.Errorf("%q is a list, cannot add sub-field %q to it (use --force to overwrite)", path, subKey)
				default:
					return fmt.Errorf("%q is not a mapping, cannot add sub-field %q to it (use --force to overwrite)", path, subKey)
				}
			}
			fmt.Printf("⚠️  --force: replacing %q with a mapping\n", path)
			// Convert in place so the node stays in the tree. Drop the old value's
			// line comment: on a mapping it gets emitted after the next key.
			valNode.Kind = yaml.MappingNode
			valNode.Tag = "!!map"
			valNode.Value = ""
			valNode.Style = 0
			valNode.Content = nil
			valNode.Alias = nil
			valNode.LineComment = ""
		}
		curr = valNode
	}
	return nil
}

// ---------------------------------------------------------
// Low-level: Searches a MappingNode's Content slice for a key
// ---------------------------------------------------------
func findKeyInMapping(mapping *yaml.Node, key string) *yaml.Node {
	if mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1] // Returns the value node (could be scalar or nested map)
		}
	}
	return nil
}

// printMapOrder renders a mapping or list as indented "key : value" / "- item" lines
func printMapOrder(node *yaml.Node, indent string) []string {
	var printContents []string
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			if valNode.Kind == yaml.MappingNode || valNode.Kind == yaml.SequenceNode {
				printContents = append(printContents, fmt.Sprintf("%s%s:\n", indent, keyNode.Value))
				printContents = append(printContents, printMapOrder(valNode, indent+"  ")...)
			} else {
				printContents = append(printContents, fmt.Sprintf("%s%s : %s\n", indent, keyNode.Value, valNode.Value))
			}
		}
	case yaml.SequenceNode:
		for _, item := range node.Content {
			if item.Kind == yaml.MappingNode || item.Kind == yaml.SequenceNode {
				printContents = append(printContents, fmt.Sprintf("%s-\n", indent))
				printContents = append(printContents, printMapOrder(item, indent+"  ")...)
			} else {
				printContents = append(printContents, fmt.Sprintf("%s- %s\n", indent, item.Value))
			}
		}
	}
	return printContents
}
