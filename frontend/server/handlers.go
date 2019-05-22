package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mruck/athena/lib/exception"
	"gopkg.in/mgo.v2"
)

type Server struct {
	Exceptions *exception.ExceptionsManager
}

func NewServer(db *mgo.Database) (*Server, error) {
	exceptions := exception.NewExceptionsManager(db)
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

	WriteJSONResponse(results, w)
}

// FuzzTargetInfo is the return result from FuzzTarget
type FuzzTargetInfo struct {
	PodName  string
	TargetID string
}

// FuzzTarget is an endpoint to upload metadata about a target and start a fuzz job
// It returns FuzzTargetInfo on success
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

	// Sanity check that the target is fuzzable
	pod, err := DryRun(&target)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Launch pod for fuzzing
	err = Fuzz(pod, &target)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Return the target id and pod name for querying later on
	w.WriteHeader(http.StatusOK)
	info := FuzzTargetInfo{TargetID: pod.ObjectMeta.Labels["TargetID"],
		PodName: pod.Name}
	WriteJSONResponse(info, w)
}
