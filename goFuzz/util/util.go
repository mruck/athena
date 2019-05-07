package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
