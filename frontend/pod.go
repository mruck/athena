package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/api/core/v1"
)

// Generate an Athena Container.
func getAthenaContainer(targetId string) v1.Container {
	var AthenaContainer = v1.Container{
		Name:    "athena",
		Image:   "gcr.io/athena-fuzzer/athena:07b1cc1e09",
		Command: []string{"./run_client.sh"},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      "results-dir",
				MountPath: "/tmp/results",
			},
		},
	}
	AthenaContainer.Env = []v1.EnvVar{
		v1.EnvVar{Name: "TARGET_ID", Value: targetId},
	}
	return AthenaContainer

}

// Add the Athena container to the uninstrumented pod
func InjectAthenaContainer(pod v1.Pod) v1.Pod {
	athenaPodName := NewTargetId()
	pod.ObjectMeta.Name = athenaPodName
	athenaContainer := getAthenaContainer(athenaPodName)
	pod.Spec.Containers = append(pod.Spec.Containers, athenaContainer)
	return pod
}

// Build a vanilla pod spec.  This pod is uninstrumented/barebones
// (i.e. no Athena container injected)
func buildPod(containers []v1.Container) v1.Pod {
	var pod v1.Pod
	// Basic initialization
	pod.APIVersion = "v1"
	pod.Kind = "Pod"
	// Unique identifier for pod and target
	targetId := NewTargetId()
	pod.ObjectMeta.Name = targetId
	pod.ObjectMeta.Labels = map[string]string{"fuzz_pod": "true", "target_id": targetId}
	// Add target containers
	pod.Spec.Containers = containers
	// Add shared mount
	pod.Spec.Volumes = []v1.Volume{
		v1.Volume{
			Name: "results-dir",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}
	return pod
}

const PodSpecDir = "/tmp/pod_specs"

// Get a file to write the pod spec to
func getPodSpecDest(pod v1.Pod) string {
	_ = os.Mkdir(PodSpecDir, 0700)
	return filepath.Join(PodSpecDir, pod.ObjectMeta.Name)
}

// Marshal pod and write to disc.
func writePodSpecToDisc(pod v1.Pod, dst string) error {

	// Marshal pod
	podBytes, err := json.Marshal(pod)
	if err != nil {
		err = fmt.Errorf("Error marshaling pod spec: %v", err)
		return err
	}

	// Write pod spec to disc
	err = ioutil.WriteFile(dst, podBytes, 0644)
	if err != nil {
		err = fmt.Errorf("Error writing pod spec to disc: %v", err)
		return err
	}

	return nil

}

// Given a v1.Pod, write the spec to disc, launch the pod, then poll
// until all containers are ready or it times out
func RunPod(w http.ResponseWriter, pod v1.Pod, deletePod bool) error {
	// Write pod spec to disc
	podSpecPath := getPodSpecDest(pod)
	err := writePodSpecToDisc(pod, podSpecPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return err
	}

	// Launch pod
	err = LaunchPod(podSpecPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return err
	}

	// Poll pod until it's ready or we hit a timeout
	ready, err := PollPodReady(pod.ObjectMeta.Name)

	// Reap the pod if specified
	if deletePod == true {
		err = DeletePod(pod.ObjectMeta.Name)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return err
		}
	}

	// Handle polling errors
	if err != nil {
		http.Error(w, err.Error(), 500)
		return err
	} else if ready != true {
		err = fmt.Errorf("Pod not ready. Are there enough resources? Maybe you should delete all pods")
		http.Error(w, err.Error(), 500)
		return err
	}

	return nil

}
