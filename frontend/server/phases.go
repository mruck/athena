package server

import (
	"fmt"
	"net/http"

	v1 "k8s.io/api/core/v1"
)

// Run an uninstrumented pod
func runVanillaPod(target *Target) (*v1.Pod, error) {
	fmt.Println("Launching vanilla pod")
	// Generate a vanilla pod with the user provided containers
	pod := buildPod(target.Containers, *target.Name)

	// Sanity check that the uninstrumented target runs
	err := RunPod(&pod, true)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

// Run pod with our rails mounted in
func runCustomRailsPod(pod *v1.Pod, target *Target) error {
	fmt.Println("Launching pod instrumented with rails")
	// Modifies the pod spec in memory to point to our rails
	InstrumentRails(pod, target)

	// Sanity check that the uninstrumented target runs
	return RunPod(pod, true)
}

// DryRun sanity checks that our target is fuzzable.
// If so, it returns a pod for the target.
func DryRun(target *Target) (*v1.Pod, error) {
	// Run the target as the user provided
	pod, err := runVanillaPod(target)
	if err != nil {
		return nil, err
	}
	// Run the target with our rails-fork
	err = runCustomRailsPod(pod, target)
	if err != nil {
		return nil, err
	}
	return pod, nil
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
