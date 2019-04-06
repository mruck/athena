package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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

func buildRailsContainer() v1.Container {
	image := os.Getenv("RAILS_IMAGE")
	if image == "" {
		image = "gcr.io/athena-fuzzer/rails:6f8a54aa0acd97a1c780"
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

//MakeFuzzable makes a pod fuzzable by injecting the Athena container
func MakeFuzzable(pod *v1.Pod, target *Target) error {
	// Make this a new pod
	name := pod.ObjectMeta.Labels["name"]
	pod.ObjectMeta.Name = NewPodID(name)

	// Unique identifier for target
	targetID := NewTargetID(name)
	pod.ObjectMeta.Labels = map[string]string{"fuzz_pod": "true", "TargetID": targetID, "name": name}

	// Mount rails-fork
	mountRails(pod)

	// Mount results directory for sharing results
	mountResultsDir(pod)

	// Add the Athena Container to the uninstrumented pod
	return InjectAthenaContainer(pod, target)
}

// Build a vanilla pod spec.  This pod is uninstrumented/barebones
// (i.e. no Athena container injected)
func buildPod(containers []v1.Container, name string) v1.Pod {
	var pod v1.Pod
	// Basic initialization
	pod.APIVersion = "v1"
	pod.Kind = "Pod"
	pod.ObjectMeta.Name = NewPodID(name)
	pod.ObjectMeta.Labels = map[string]string{"name": name}
	// Add target containers
	pod.Spec.Containers = containers
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
		err = fmt.Errorf("error marshaling pod spec: %v", err)
		return err
	}

	// Write pod spec to disc
	err = ioutil.WriteFile(dst, podBytes, 0644)
	if err != nil {
		err = fmt.Errorf("error writing pod spec to disc: %v", err)
		return err
	}
	fmt.Printf("Pod spec written to %s\n", dst)

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
