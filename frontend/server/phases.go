package server

import (
	"net/http"

	v1 "k8s.io/api/core/v1"
)

// Run an uninstrumented pod
func runVanillaPod(target *Target) (*v1.Pod, error) {
	// Generate a vanilla pod with the user provided containers
	pod := buildPod(target.Containers, *target.Name)

	// Sanity check that the uninstrumented target runs
	err := RunPod(&pod, true)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

// DryRun sanity checks that our target is fuzzable.
// If so, it returns a pod for the target.
func DryRun(target *Target) (*v1.Pod, error) {
	// Run the target as the user provided
	return runVanillaPod(target)
	// Run the target with our rails-fork
}

// Fuzz takes a pod, makes it fuzzable, and launches it.
// Upon success, it writes the target id to the response
// for querying by the client.
func Fuzz(pod *v1.Pod, target *Target, w http.ResponseWriter) {
	err := MakeFuzzable(pod, target)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Launch the pod with the athena container
	err = RunPod(pod, false)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Return the target id for querying later on
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(pod.ObjectMeta.Labels["TargetID"]))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}