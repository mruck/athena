package server

import (
	"github.com/mruck/athena/frontend/log"
	v1 "k8s.io/api/core/v1"
)

// Run an uninstrumented pod
func runVanillaPod(target *Target) (*v1.Pod, error) {
	log.Info("\nLaunching vanilla pod")
	// Generate a vanilla pod with the user provided containers
	pod := buildPod(target.Containers, *target.Name)
	defer DeletePod(pod.ObjectMeta.Name)

	// Sanity check that the uninstrumented target runs
	err := RunPod(&pod)
	if err != nil {
		return nil, err
	}
	// TODO: readiness probe
	return &pod, nil
}

// Run pod with our rails mounted in
func runCustomRailsPod(pod *v1.Pod, target *Target) error {
	log.Info("\nLaunching pod instrumented with rails")
	// Modifies the pod spec in memory to point to our rails
	InstrumentRails(pod, target)
	defer DeletePod(pod.ObjectMeta.Name)

	// Sanity check that the uninstrumented target runs
	err := RunPod(pod)
	if err != nil {
		return err
	}
	// TODO: readiness probe
	return nil
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
func Fuzz(pod *v1.Pod, target *Target) error {
	err := MakeFuzzable(pod, target)
	if err != nil {
		return err
	}

	// Launch the pod with the athena container
	return RunPod(pod)
}
