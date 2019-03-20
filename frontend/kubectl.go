package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

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
	fmt.Println(string(stdoutStderr))
	return nil

}

func GetContainerStatuses(podName string) ([]v1.ContainerStatus, error) {
	cmd := exec.Command("kubectl", "get", "pod", podName, "-o", "json")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	if cmd.ProcessState.ExitCode() != 0 {
		return nil, err
	}
	var pod v1.Pod
	err = json.Unmarshal(stdoutStderr, &pod)
	if err != nil {
		return nil, err
	}
	return pod.Status.ContainerStatuses, nil

}

func PodReady(containerStatuses []v1.ContainerStatus) bool {
	for _, containerStatus := range containerStatuses {
		if containerStatus.Ready != true {
			fmt.Printf("Checking %v container...Ready : %v\n", containerStatus.Name, containerStatus.Ready)
			return false
		}
	}
	return true
}

// Check if each container in the pod is ready, i.e.
// each ContainerStatus.Ready = True in the pod.Status.ContainerStatuses array
func PollPodReady(podName string) (bool, error) {
	for start := time.Now(); time.Since(start) < 120*time.Second; {
		print("Sleeping...")
		time.Sleep(5 * time.Second)
		containerStatuses, err := GetContainerStatuses(podName)
		if err != nil {
			return false, err
		}
		ready := PodReady(containerStatuses)
		if ready == true {
			return true, nil
		}
	}
	return false, nil
}
