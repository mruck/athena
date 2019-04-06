package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	v1 "k8s.io/api/core/v1"
)

type Server struct {
	Exceptions *ExceptionsManager
}

func NewServer(db *mgo.Database) (*Server, error) {
	exceptions := NewExceptionsManager(db)
	return &Server{Exceptions: exceptions}, nil
}

func (server *Server) getRoutes() Routes {
	return Routes{
		Route{
			"Index",
			"GET",
			"/",
			server.Index,
		},
		Route{
			"Exceptions",
			"GET",
			"/Exceptions/{targetID}",
			server.ExceptionsHandler,
		},
		Route{
			"FuzzTarget",
			"POST",
			"/FuzzTarget",
			server.FuzzTarget,
		},
	}
}

// Index provides a sanity check that server is running
func (server *Server) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

//ExceptionsHandler endpoint retunrs exceptions associated with fuzz target id
func (server *Server) ExceptionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars["targetID"]
	fmt.Printf("Target id: %v", targetID)

	results, err := server.Exceptions.GetAll(targetID)
	if err != nil {
		err = fmt.Errorf("error connecting to db: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	resultBytes, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(resultBytes)
}

func runVanillaPod(containers []v1.Container, target *Target) (*v1.Pod, error) {
	// Generate a vanilla pod with the user provided containers
	pod := buildPod(target.Containers, *target.Name)

	// Sanity check that the uninstrumented target runs
	err := RunPod(&pod, true)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

// FuzzTarget is an endpoint to upload metadata about a target and start a fuzz job
func (server Server) FuzzTarget(w http.ResponseWriter, r *http.Request) {
	// Get list of containers pushed by user
	var target Target
	err := ParseBody(w, r, &target)
	if err != nil {
		return
	}

	// Ensure user provided data is valid
	err = ValidateTarget(&target)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	pod, err := runVanillaPod(target.Containers, &target)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = MakeFuzzable(pod, &target)
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

	// We are fuzzing!
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pod.ObjectMeta.Labels["TargetID"]))
}
