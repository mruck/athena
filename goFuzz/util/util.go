package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
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

// MustUnmarshalFile reads a file and unmarshal it to the given
// destination, panicking on error
func MustUnmarshalFile(filepath string, dst interface{}) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		err = errors.Wrap(err, "")
		log.Fatalf("%+v\n", err)
	}
	err = json.Unmarshal(data, dst)
	if err != nil {
		err = errors.Wrap(err, "")
		log.Fatalf("%+v\n", err)
	}
}
