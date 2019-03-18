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
		http.Error(w, "Error reading body:", 500)
		http.Error(w, err.Error(), 500)
		return nil, err
	}
	// Unmarshal
	var containers []v1.Container
	err = json.Unmarshal(b, &containers)
	if err != nil {
		http.Error(w, "Error unmarshaling []v1.Container:", 500)
		http.Error(w, err.Error(), 500)
		return nil, err
	}
	return containers, nil
}

func PushPod(w http.ResponseWriter, r *http.Request) {
	containers, err := readBody(w, r)
	if err != nil {
		return
	}

	pod := buildPod(containers)

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
