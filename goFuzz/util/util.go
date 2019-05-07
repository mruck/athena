package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// PrettyPrint a struct
func PrettyPrint(data interface{}) {
	jsonified, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		err = fmt.Errorf("failed to pretty print json: %v", err)
		log.Fatal(err)
	}
	fmt.Println(string(jsonified))
}

// PatchRequestHostPort replaces the host and port in an http.request
func PatchRequestHostPort(request *http.Request, host string, port string) {
	request.URL.Host = host + ":" + port
	request.Host = host + ":" + port
}

// PatchRequestsHostPort replaces the host and port in a list of requests
func PatchRequestsHostPort(requests []*http.Request, host string, port string) {
	for _, req := range requests {
		PatchRequestHostPort(req, host, port)
	}
}

// MustGetTargetAppPort returns the port the target app is running on
// or panics
func MustGetTargetAppPort() string {
	port := os.Getenv("TARGET_APP_PORT")
	if port == "" {
		log.Fatal("TARGET_APP_PORT not set")
	}
	return port
}

// MustGetTargetAppHost returns the host of the target app or localhost
// if not set
func MustGetTargetAppHost() string {
	host := os.Getenv("TARGET_APP_HOST")
	if host == "" {
		return "localhost"
	}
	return host
}
