package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

// PanicIfErr panics and logs a verbose error if err is not nil.
func PanicIfErr(err error) {
	if err != nil {
		err = errors.Wrap(err, "")
		log.Fatalf("%+v\n", err)
	}
}

// Verbose makes an error verbose.
func Verbose(err error) error {
	return errors.Wrap(err, "")
}

// ReadLines reads a file line by line
func ReadLines(file io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "")
	}
	return lines, nil
}
