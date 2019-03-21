package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/api/core/v1"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

const DbName = "athena"
const Localhost = "localhost"
const Port = "27101"

// Return exceptions associated with fuzz target id
func Exceptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetId := vars["targetId"]
	fmt.Fprintf(w, "Target id: %v", targetId)
	// Poll mongodb
	client, err := NewClient(Host, Port, DbName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	results, err := client.LookUp(targetId)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println(results)
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

func FuzzTarget(w http.ResponseWriter, r *http.Request) {
	// Get list of containers pushed by user
	containers, err := readBody(w, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Generate a vanilla pod with the user provided containers
	pod := buildPod(containers)

	// Sanity check that the uninstrumented target runs
	err = RunPod(w, pod, true)
	if err != nil {
		return
	}

	// Add the Athena Container to the uninstrumented pod
	pod = InjectAthenaContainer(pod)

	// Launch the pod with the athena container
	err = RunPod(w, pod, false)
	if err != nil {
		return
	}

	// We are fuzzing!
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pod.ObjectMeta.Name))
}
