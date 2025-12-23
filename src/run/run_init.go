package run   // 👈 MUST match folder name

import (
	"os"
	"os/exec"
)


func RunInit(args []string) error {


    cmd := exec.Command("python3",
        append([]string{"src/run/run_deps/run.py"}, args...)...,
    )

    // Attach stdio so output shows in terminal
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    return cmd.Run()
}