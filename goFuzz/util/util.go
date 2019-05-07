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

// PatchHostPort replaces the host and port in an http.request
func PatchHostPort(request *http.Request, host string, port string) {
	request.URL.Host = host + ":" + port
}
