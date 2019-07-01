package util

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/moul/http2curl"
	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
)

// StringInSlice iterates over the list, checking if `a` contains
// any element from the list. i.e. `a` maybe a full blown error string
// like `failed to COPY from STDIN` and the list only has keywords
// like `COPY`
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(a, b) {
			return true
		}
	}
	return false
}

// PrettyPrintStruct prints a struct as info level by default
// unless a different logging level function is passed in
func PrettyPrintStruct(data interface{}, logFn log.Fn) {
	if logFn == nil {
		logFn = log.Infof
	}
	jsonified, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		err = fmt.Errorf("failed to pretty print json: %v", err)
		log.Warn(err)
		return
	}
	logFn(string(jsonified))
}

// GetLogPath returns where custom Athena data should be stored,
// i.e. athena errors, parsed sql errors, etc
func GetLogPath() string {
	path := os.Getenv("ATHENA_LOG_PATH")
	if path == "" {
		if runtime.GOOS == "darwin" {
			// Development log path, i.e. /tmp on osx
			return log.DevPath
		}
		// Production log path, i.e. /var/log/athena on k8s
		return log.Path
	}
	return path
}

// DefaultEnv returns the environment variable <key> if it is found,
// and _default otherwise.
func DefaultEnv(key string, _default string) string {
	val, found := os.LookupEnv(key)
	if !found {
		return _default
	}
	return val
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
		return false, errors.WithStack(err)
	}
	defer fp.Close()
	result, err := fp.Stat()
	if err != nil {
		return false, errors.WithStack(err)
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

// LoadCSVFile loads a cvs file at `path`
func LoadCSVFile(path string) ([][]string, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	reader := csv.NewReader(fp)
	records, err := reader.ReadAll()
	return records, errors.WithStack(err)
}

// CopyFile takes in the path to the dst and src. The dst is truncated.
func CopyFile(dst string, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()

}

// PrintType prints the type of data
func PrintType(data interface{}) {
	log.Infof("Type = %T", data)
}

// ReadFileLineByLine opens a file and reads it line by line
func ReadFileLineByLine(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ToCurl logs a request as a curl cmd
func ToCurl(req *http.Request) (*http2curl.CurlCommand, error) {
	cmd, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return cmd, nil
}
