package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/google/uuid"
)

//NewTargetID returns a uuid for the fuzz target of the form:
// targetName - target - uuid
func NewTargetID(name string) string {
	return name + "-target-" + uuid.New().String()[:8]
}

//NewPodID returns a uuid for the pod of the form:
// targetName - pod - uuid
func NewPodID(name string) string {
	return name + "-pod-" + uuid.New().String()[:8]
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

// ParseBody reads from request and marshal into opaque struct
func ParseBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading from body: %v", err)
		http.Error(w, err.Error(), 500)
		return err
	}
	// Unmarshal
	err = json.Unmarshal(b, dst)
	if err != nil {
		err = fmt.Errorf("error unmarshaling []v1.Container: %v", err)
		http.Error(w, err.Error(), 500)
		return err
	}
	return nil
}
