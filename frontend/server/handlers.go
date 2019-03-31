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

// TargetDB containers metadata about the database we are instrumenting
type TargetDB struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
}

// Target is the expected form of user input
type Target struct {
	// Name of the target application
	Name string
	// Port the target app is running on
	Port       int
	Db         TargetDB
	Containers []v1.Container
}

// FuzzTarget is an endpoint to upload metadata about a target and start a fuzz job
func (server Server) FuzzTarget(w http.ResponseWriter, r *http.Request) {
	// Get list of containers pushed by user
	var target Target
	err := ParseBody(w, r, &target)
	if err != nil {
		return
	}

	// Generate a vanilla pod with the user provided containers
	pod := buildPod(target.Containers, target.Name)

	// Sanity check that the uninstrumented target runs
	err = RunPod(w, pod, true)
	if err != nil {
		return
	}

	err = MakeFuzzable(&pod, &target)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Launch the pod with the athena container
	err = RunPod(w, pod, false)
	if err != nil {
		return
	}

	// We are fuzzing!
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pod.ObjectMeta.Labels["TargetID"]))
}
