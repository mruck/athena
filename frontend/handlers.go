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

func PushPod(w http.ResponseWriter, r *http.Request) {
	// Read from body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Unmarshal
	var pod v1.Pod
	err = json.Unmarshal(b, &pod)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Modify in place
	pod.Spec.Containers = append(pod.Spec.Containers, AthenaContainer)

	// Randomly generate name
	pod.ObjectMeta.Name = uuid.New().String()

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
