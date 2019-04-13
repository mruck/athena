package server

import (
	"fmt"
	"os"
	"strconv"

	v1 "k8s.io/api/core/v1"
)

const resultsPath = "/tmp/results"

func buildEnv(targetID string, target *Target) []v1.EnvVar {
	return []v1.EnvVar{
		v1.EnvVar{Name: "TARGET_APP_PORT", Value: strconv.Itoa(*target.Port)},
		v1.EnvVar{Name: "TARGET_DB_HOST", Value: *target.Db.Host},
		v1.EnvVar{Name: "TARGET_DB_USER", Value: *target.Db.User},
		v1.EnvVar{Name: "TARGET_DB_PASSWORD", Value: *target.Db.Password},
		v1.EnvVar{Name: "TARGET_DB_PORT", Value: strconv.Itoa(*target.Db.Port)},
		v1.EnvVar{Name: "TARGET_DB_NAME", Value: *target.Db.Name},
		v1.EnvVar{Name: "TARGET_ID", Value: targetID},
		v1.EnvVar{Name: "RESULTS_PATH", Value: resultsPath},
	}
}

// Generate an Athena Container.
func buildAthenaContainer(targetID string, target *Target) (*v1.Container, error) {
	// We expect the frontend to be provided the latest athena image,
	// otherwise how else will we know which image to run?
	image := os.Getenv("ATHENA_IMAGE")
	if image == "" {
		err := fmt.Errorf("no Athena image provided")
		return nil, err
	}
	env := buildEnv(targetID, target)
	var AthenaContainer = v1.Container{
		Name:  "athena",
		Image: image,
		Env:   env,
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      "results-dir",
				MountPath: "/tmp/results",
			},
		},
	}
	return &AthenaContainer, nil

}

// Add the Athena container to the uninstrumented pod
func InjectAthenaContainer(pod *v1.Pod, target *Target) error {
	targetID := pod.ObjectMeta.Labels["TargetID"]
	athenaContainer, err := buildAthenaContainer(targetID, target)
	if err != nil {
		return err
	}
	pod.Spec.Containers = append(pod.Spec.Containers, *athenaContainer)
	return nil
}

// InjectSideCar adds a copy of the athena container that just sleeps for debugging
func InjectSideCar(pod *v1.Pod, target *Target) error {
	// Get an Athena Container
	targetID := pod.ObjectMeta.Labels["TargetID"]
	athenaContainer, err := buildAthenaContainer(targetID, target)
	if err != nil {
		return err
	}
	athenaContainer.Name = "sidecar-athena"
	athenaContainer.Command = []string{"/bin/bash"}
	athenaContainer.Args = []string{"-c", "while true; do sleep 1000; done"}
	pod.Spec.Containers = append(pod.Spec.Containers, *athenaContainer)
	return nil
}

func buildRailsContainer() v1.Container {
	image := os.Getenv("RAILS_IMAGE")
	if image == "" {
		image = "gcr.io/athena-fuzzer/rails:dafb06189a5efeaefc32"
		fmt.Printf("No rails image provided.  Using default image %s\n", image)
	}
	return v1.Container{
		Name:  "rails-fork",
		Image: image,
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      "rails-fork",
				MountPath: "/rails-fork",
			},
		},
	}
}

func GetTargetContainer(containers []v1.Container) *v1.Container {
	for i, container := range containers {
		// Found it
		if container.Name == "target" {
			return &containers[i]
		}
	}
	return nil
}

// Build init container for rails-fork and add shared mount in target directory
// to copy rails-fork to
func mountRails(pod *v1.Pod) {
	// Generate rails-fork container
	railsContainer := buildRailsContainer()
	pod.Spec.InitContainers = []v1.Container{railsContainer}

	// Add rails-fork mount point to target container so that it can use our rails
	railsVolumeMount := v1.VolumeMount{
		Name:      "rails-fork",
		MountPath: "/rails-fork",
	}
	targetContainer := GetTargetContainer(pod.Spec.Containers)
	targetContainer.VolumeMounts = append(targetContainer.VolumeMounts, railsVolumeMount)

	// Add rails-fork volume to pod spec
	railsVolume := v1.Volume{
		Name: "rails-fork",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, railsVolume)
}

// Mount results directory in for sharing results between athena and target
func mountResultsDir(pod *v1.Pod) {
	// Add results dir container
	resultsVolumeMount := v1.VolumeMount{
		Name:      "results-dir",
		MountPath: "/tmp/results",
	}
	targetContainer := GetTargetContainer(pod.Spec.Containers)
	targetContainer.VolumeMounts = append(targetContainer.VolumeMounts, resultsVolumeMount)

	// Add env var telling rails where to write results
	resultsEnvVar := v1.EnvVar{Name: "RESULTS_PATH", Value: resultsPath}
	targetContainer.Env = append(targetContainer.Env, resultsEnvVar)

	// Add results volume to pod spec
	resultsVolume := v1.Volume{
		Name: "results-dir",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, resultsVolume)
}

// InstrumentRails mounts our rails in the target container and mounts the
// results directory.  This is done in memory.
func InstrumentRails(pod *v1.Pod, target *Target) {
	// Make this a new pod
	name := pod.ObjectMeta.Labels["name"]
	pod.ObjectMeta.Name = NewPodID(name)

	// Mount rails-fork
	mountRails(pod)

	// Mount results directory for sharing results
	mountResultsDir(pod)

	// Tell our rails-fork where the target app lives
	targetContainer := GetTargetContainer(pod.Spec.Containers)
	targetContainer.Env = append(targetContainer.Env, v1.EnvVar{Name: "TARGET_APP_PATH", Value: *target.AppPath})
	targetContainer.Env = append(targetContainer.Env, v1.EnvVar{Name: "RAILS_FORK", Value: "1"})
}

//MakeFuzzable makes a pod fuzzable by injecting the Athena container
func MakeFuzzable(pod *v1.Pod, target *Target) error {
	// Make this a new pod
	name := pod.ObjectMeta.Labels["name"]
	pod.ObjectMeta.Name = NewPodID(name)

	// Unique identifier for target
	targetID := NewTargetID(name)
	pod.ObjectMeta.Labels = map[string]string{"fuzz_pod": "true", "TargetID": targetID, "name": name}

	// Add the Athena Container to the uninstrumented pod
	err := InjectAthenaContainer(pod, target)
	if err != nil {
		return err
	}

	// Add a side car for debugging
	return InjectSideCar(pod, target)
}
