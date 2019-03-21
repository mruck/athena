package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"k8s.io/api/core/v1"
)

// Delete pod via kubectl
func DeletePod(podName string) error {
	// Launch pod
	cmd := exec.Command("kubectl", "delete", "pod", podName)
	_, err := ExecWrapper(cmd)
	return err
}

//Spin up pod with kubectl exec
func LaunchPod(podSpecPath string) error {
	// Launch pod
	cmd := exec.Command("kubectl", "apply", "-f", podSpecPath)
	_, err := ExecWrapper(cmd)
	return err
}

// Parse JSON dump for container status
func GetContainerStatuses(podName string) ([]v1.ContainerStatus, error) {
	cmd := exec.Command("kubectl", "get", "pod", podName, "-o", "json")
	stdoutStderr, err := ExecWrapper(cmd)
	if err != nil {
		return nil, err
	}
	var pod v1.Pod
	err = json.Unmarshal(stdoutStderr, &pod)
	if err != nil {
		return nil, err
	}
	return pod.Status.ContainerStatuses, nil

}

// Check if each container in the pod is ready, i.e.
// each ContainerStatus.Ready = True in the pod.Status.ContainerStatuses array
func PodReady(containerStatuses []v1.ContainerStatus) bool {
	if len(containerStatuses) == 0 {
		return false
	}
	for _, containerStatus := range containerStatuses {
		fmt.Printf("Checking %v container...Ready : %v\n", containerStatus.Name, containerStatus.Ready)
		if containerStatus.Ready != true {
			return false
		}
	}
	return true
}

// Poll pods to check if they are ready, with a 120s timeout
func PollPodReady(podName string) (bool, error) {
	for start := time.Now(); time.Since(start) < 120*time.Second; {
		time.Sleep(5 * time.Second)
		containerStatuses, err := GetContainerStatuses(podName)
		if err != nil {
			fmt.Println("Failed to get container status", err)
			return false, err
		}
		ready := PodReady(containerStatuses)
		if ready == true {
			return true, nil
		}
	}
	fmt.Println("Pod not ready. Are there enough resources? Maybe you should delete all pods", podName)
	return false, nil
}
