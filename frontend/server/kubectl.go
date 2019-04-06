package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
)

// Delete pod via kubectl and log any errors
func DeletePod(podName string) {
	// Launch pod
	cmd := exec.Command("kubectl", "delete", "pod", podName)
	_, err := ExecWrapper(cmd)
	if err != nil {
		//log.Errof("Failed to delete pod %s: %v", podName, err)
		fmt.Printf("Failed to delete pod %s: %v", podName, err)
	}
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

// Build a vanilla pod spec.  This pod is uninstrumented/barebones
// (i.e. no Athena container injected)
func buildPod(containers []v1.Container, name string) v1.Pod {
	var pod v1.Pod
	// Basic initialization
	pod.APIVersion = "v1"
	pod.Kind = "Pod"
	pod.ObjectMeta.Name = NewPodID(name)
	pod.ObjectMeta.Labels = map[string]string{"name": name}
	// Add target containers
	pod.Spec.Containers = containers
	return pod
}

const PodSpecDir = "/tmp/pod_specs"

// Get a file to write the pod spec to
func getPodSpecDest(pod *v1.Pod) string {
	_ = os.Mkdir(PodSpecDir, 0700)
	return filepath.Join(PodSpecDir, pod.ObjectMeta.Name)
}

// Marshal pod and write to disc.
func writePodSpecToDisc(pod *v1.Pod, dst string) error {

	// Marshal pod
	podBytes, err := json.Marshal(*pod)
	if err != nil {
		err = fmt.Errorf("error marshaling pod spec: %v", err)
		return err
	}

	// Write pod spec to disc
	err = ioutil.WriteFile(dst, podBytes, 0644)
	if err != nil {
		err = fmt.Errorf("error writing pod spec to disc: %v", err)
		return err
	}
	fmt.Printf("Pod spec written to %s\n", dst)

	return nil

}

// Given a v1.Pod, write the spec to disc, launch the pod, then poll
// until all containers are ready or it times out
func RunPod(pod *v1.Pod) error {
	// Write pod spec to disc
	podSpecPath := getPodSpecDest(pod)
	err := writePodSpecToDisc(pod, podSpecPath)
	if err != nil {
		return err
	}

	// Launch pod
	err = LaunchPod(podSpecPath)
	if err != nil {
		return err
	}

	// Poll pod until it's ready or we hit a timeout
	ready, err := PollPodReady(pod.ObjectMeta.Name)

	// Handle polling errors
	if err != nil {
		return err
	} else if !ready {
		err = fmt.Errorf("pod not ready. Are there enough resources? Maybe you should delete all pods")
		return err
	}

	return nil

}
