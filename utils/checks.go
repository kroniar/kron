package utils

import (
	"fmt"
	"os/exec"
)

// EnsureDependency checks if a dependency exists and suggests installation if missing
func EnsureDependency(name string) bool {
	_, err := exec.LookPath(name)
	if err != nil {
		fmt.Printf("%s is not installed ❌\n", name)
		fmt.Printf("Run 'kron setup %s' to install it.\n", name)
		return false
	}
	return true
}
