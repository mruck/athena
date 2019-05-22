package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
)

// PrettyPrintStruct prints a struct
func PrettyPrintStruct(data interface{}) {
	jsonified, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		err = fmt.Errorf("failed to pretty print json: %v", err)
		log.Fatal(err)
	}
	fmt.Println(string(jsonified))
}

// MustGetTargetID returns the target id or panics
func MustGetTargetID() string {
	targetID := os.Getenv("TARGET_ID")
	if targetID == "" {
		log.Fatal("TARGET_ID not set")
	}
	return targetID
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

// FileIsEmpty returns whether or not a file is empty
func FileIsEmpty(filepath string) (bool, error) {
	fp, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer fp.Close()
	result, err := fp.Stat()
	if err != nil {
		return false, err
	}
	return result.Size() == 0, nil
}

// UnmarshalFile reads a file and unmarshal it to the given
// destination, returning the error
func UnmarshalFile(filepath string, dst interface{}) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.WithStack(err)
	}
	err = json.Unmarshal(data, dst)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// MustUnmarshalFile reads a file and unmarshals it to the given
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

// Must requires check to succeed otherwise panic
func Must(check bool, format string, args ...interface{}) {
	if !check {
		log.Fatalf(format, args...)
	}
}

// MarshalToFile marshal a struct to a file
func MarshalToFile(data interface{}, dst string) error {
	JSONData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, JSONData, 0644)
}

// Stringify takes an arbitrary primitive type and converts it to a string
// Note:  doesn't support arrays or objects!
// TODO: figure out how to support arrays
func Stringify(data interface{}) string {
	return fmt.Sprintf("%v", data)
}
