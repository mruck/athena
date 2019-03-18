package main

import (
	"fmt"
	"os/exec"
)

func launchPod(podSpecPath string) error {
	// Launch pod
	cmd := exec.Command("kubectl", "apply", "-f", podSpecPath)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if cmd.ProcessState.ExitCode() != 0 {
		err = fmt.Errorf("Error spawning pod: %v", stdoutStderr)
		return err
	}
	fmt.Printf("%s\n", stdoutStderr)
	return nil

}
