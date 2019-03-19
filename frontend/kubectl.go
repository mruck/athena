package main

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"k8s.io/api/core/v1"
)

//Sin up pod with kubectl exec
func LaunchPod(podSpecPath string) error {
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

func GetStatus(podName string) error {
	cmd := exec.Command("kubectl", "get", "pod", podName, "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if cmd.ProcessState.ExitCode() != 0 {
		return err
	}
	var pod v1.Pod
	err = json.Unmarshal(stdoutStderr, &pod)
	if err != nil {
		return err
	}
	var containerStatuses []v1.ContainerStatus
	containerStatuses = pod.Status.ContainerStatuses
	for _, containerStatus := range containerStatuses {
		fmt.Println(containerStatus.Name)
		fmt.Println(containerStatus.Ready)
	}
	return nil

}
