package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	"github.com/google/uuid"
	"k8s.io/api/core/v1"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func getAthenaContainer(targetId string) v1.Container {
	var AthenaContainer = v1.Container{
		Name:    "athena",
		Image:   "gcr.io/athena-fuzzer/athena:0c17b6038c",
		Command: []string{"./run_client.sh"},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      "postgres-socket",
				MountPath: "/var/run/postgresql",
			},
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
	return pod
}

func PushPod(w http.ResponseWriter, r *http.Request) {
	// Read from body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Unmarshal
	var containers []v1.Container
	err = json.Unmarshal(b, &containers)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var pod v1.Pod
	// Modify in place
	pod.Spec.Containers = containers
	pod.Spec.Containers = append(pod.Spec.Containers, AthenaContainer)

	// Dump to disc
	podBytes, err := json.Marshal(pod)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("/tmp/marli_pod.json", podBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Launch pod
	cmd := exec.Command("kubectl", "apply", "-f", "/tmp/marli_pod.json")
	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)

	// TODO: some sort of health check
}
