package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	var result Exception
	result, err := server.Exceptions.ReadOne(targetID)
	if err != nil {
		err = fmt.Errorf("error connecting to db: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Println(result)
}

// Read in user data.  We expect: a target name, []v1.Container, a database name, type
// and port.
func readBody(w http.ResponseWriter, r *http.Request) ([]v1.Container, error) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading from body: %v", err)
		http.Error(w, err.Error(), 500)
		return nil, err
	}
	// Unmarshal
	var containers []v1.Container
	err = json.Unmarshal(b, &containers)
	if err != nil {
		err = fmt.Errorf("error unmarshaling []v1.Container: %v", err)
		http.Error(w, err.Error(), 500)
		return nil, err
	}
	return containers, nil
}

func (server Server) FuzzTarget(w http.ResponseWriter, r *http.Request) {
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
