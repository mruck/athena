package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
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
	_, err = w.Write(resultBytes)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
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

	// Sanity check that the target is fuzzable
	pod, err := DryRun(&target)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	Fuzz(pod, &target, w)
}
