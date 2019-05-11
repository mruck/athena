package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-openapi/spec"
	"github.com/google/uuid"
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

// RandInt returns a truncated uuid
func RandInt() uint32 {
	uid := uuid.New()
	return uid.ID()
}

// RandString returns a stringified uuid
func RandString() string {
	uid := uuid.New()
	return uid.String()
}

// RandBool returns true or false
// TODO: make this actually random!
func RandBool() bool {
	return true
}

// RandDecimal returns true or false
func RandDecimal() float32 {
	return float32(RandInt()) / 100
}

// Rand returns a random object of type typ
// From god docs on spec library:
// type StringOrArray []string
// StringOrArray represents a value that can either be a string or an array of
// strings. Mainly here for serialization purposes
func Rand(typ spec.StringOrArray) interface{} {
	// TODO: use a rng seeded with 0 for reproducability?
	dataType := typ[0]
	if dataType == "string" {
		return RandString()
	}
	if dataType == "int" {
		return RandInt()
	}
	if dataType == "number" {
		return RandInt()
	}
	if dataType == "boolean" {
		return RandBool()
	}
	if dataType == "decimal" {
		return RandDecimal()
	}
	fmt.Printf("Unsupport data type: %s\n", dataType)
	return RandString()
}
