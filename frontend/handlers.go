package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"net/http"
	"os"
	"path/filepath"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

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

// Build a vanilla pod spec.
func buildPod(containers []v1.Container) v1.Pod {
	var pod v1.Pod
	// Basic initialization
	pod.APIVersion = "v1"
	pod.Kind = "Pod"
	// Use this as the target id
	targetId := uuid.New().String()
	pod.ObjectMeta.Name = targetId
	pod.ObjectMeta.Labels = map[string]string{"fuzz_pod": "true", "target_id": targetId}
	// Add target containers
	pod.Spec.Containers = containers
	// Inject Athena container
	athenaContainer := getAthenaContainer(targetId)
	pod.Spec.Containers = append(pod.Spec.Containers, athenaContainer)
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

// Read in user data.  We expect: a target name, []v1.Container, a database name, type
// and port.
func readBody(w http.ResponseWriter, r *http.Request) ([]v1.Container, error) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		err = fmt.Errorf("Error reading from body: %v", err)
		http.Error(w, err.Error(), 500)
		return nil, err
	}
	// Unmarshal
	var containers []v1.Container
	err = json.Unmarshal(b, &containers)
	if err != nil {
		err = fmt.Errorf("Error unmarshaling []v1.Container: %v", err)
		http.Error(w, err.Error(), 500)
		return nil, err
	}
	return containers, nil
}

const PodSpecDir = "/tmp/pod_specs"

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

func PushPod(w http.ResponseWriter, r *http.Request) {
	containers, err := readBody(w, r)
	if err != nil {
		return
	}

	pod := buildPod(containers)

	podSpecPath := getPodSpecDest(pod)
	err = writePodSpecToDisc(pod, podSpecPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = launchPod(podSpecPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	// Health check
}
