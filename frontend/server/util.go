package server

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/google/uuid"
)

// Use this as the target id
func NewTargetID() string {
	return uuid.New().String()

}

//MustGetHost returns the host platform. Useful to tell if we are on k8s or local
func MustGetHost() string {
	if runtime.GOOS == "linux" {
		return "mongodb-service"
	}
	if runtime.GOOS == "darwin" {
		return "localhost"
	}
	panic("Unsupported OS")
}

// Capture stdout/stderr of exec.Command. If it errors, wrap the error
// with stdout/stderr.  Otherwise, return stdout/stderr.
func ExecWrapper(proc *exec.Cmd) ([]byte, error) {
	stdoutStderr, err := proc.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Command run: %s\n\nStdout/stderr:  %v\nerror: %v",
			proc.Args, string(stdoutStderr), err)
		return stdoutStderr, err
	}
	if proc.ProcessState.ExitCode() != 0 {
		err = fmt.Errorf("Command run: %s\n\nStdout/stderr:  %v\nexit code: %v",
			proc.Args, string(stdoutStderr), proc.ProcessState.ExitCode())
		return stdoutStderr, err
	}
	return stdoutStderr, nil
}
