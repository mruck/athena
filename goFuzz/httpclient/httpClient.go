package httpclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
)

const maxAttempts = 60

// Try every 2 seconds
const interval = 2
const healthCheckRoute = "/rails/info/pluralization"

// Client is an http client with a new HealthCheck method defined.
type Client struct {
	*http.Client

	URL             *url.URL
	HealthcheckPath string
	StatusCodes     map[int]int
}

// New allocates an http client with a cookie jar.
func New(url *url.URL) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{Jar: jar}
	return &Client{
		Client:          httpClient,
		URL:             url,
		HealthcheckPath: healthCheckRoute,
		StatusCodes:     map[int]int{},
		// TODO: same thing with interval field that takes default
		// from a constant.
	}, nil
}

// HealthCheck checks if a hard coded rails fork endpoint is up
func (cli *Client) HealthCheck() (bool, error) {
	url := fmt.Sprintf("%s%s", cli.URL, cli.HealthcheckPath)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, errors.WithStack(err)
	}

	for i := 0; i < maxAttempts; i++ {
		log.Infof("Polling %v\n", url)
		resp, err := cli.Do(request)
		// We errored out on our last attempt
		if err != nil && i == maxAttempts-1 {
			return false, err
		}
		// Target app is up
		// TODO: the 404 status code is wonky.  Change rails fork endpoint to
		// return 200
		if err == nil && resp.StatusCode == 404 {
			return true, nil
		}
		time.Sleep(time.Second * interval)
	}

	// We never got a heartbeat
	return false, nil
}

// updateStatusCodes keeps track of status code of every response
// after sending a request
func (cli *Client) updateStatusCodes(code int) {
	if _, ok := cli.StatusCodes[code]; ok {
		cli.StatusCodes[code]++
	} else {
		cli.StatusCodes[code] = 1
	}
}

// Do will MUTATE the URL of the request passed in to have the host:port
// that the client points to. All other fields of the request
// remain intact.
func (cli *Client) Do(req *http.Request) (*http.Response, error) {
	// Patch headers
	req.Header.Add("Content-type", "application/json")
	req.Host = cli.URL.Host
	req.URL.Host = cli.URL.Host

	resp, err := cli.Client.Do(req)

	// Check if we want to log this as a curl cmd
	if _, found := os.LookupEnv("CURL"); found {
		cmd, err := util.ToCurl(req)
		if err != nil {
			log.Error(err)
		} else {
			log.Info(cmd)
		}
	}

	// Only update status codes if we got a response
	if err == nil {
		cli.updateStatusCodes(resp.StatusCode)
	}
	return resp, errors.WithStack(err)
}

// DoAll calls `.Do` on all requests and returns the first non-nil error
// or nil if they all succeed.
func (cli *Client) DoAll(requests []*http.Request) error {
	for _, request := range requests {
		_, err := cli.Do(request)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}
	return nil
}

// PrettyPrintRequest pretty prints http request at log level passed in
// or info if nil
func PrettyPrintRequest(req *http.Request, logFn log.Fn) {
	if logFn == nil {
		logFn = log.Infof
	}
	url := req.URL.Path
	if req.Method == "GET" {
		if req.URL.RawQuery != "" {
			url += "?" + req.URL.RawQuery
		}
	}
	logFn("%v %v", req.Method, url)
	if req.Body != nil {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Warnf("failed to read request: %+v", errors.WithStack(err))
			return
		}
		dst := map[string]interface{}{}
		err = json.Unmarshal(b, &dst)
		if err != nil {
			log.Warnf("%+v", errors.WithStack(err))
		}
		util.PrettyPrintStruct(dst, logFn)
		// Reinitialize the reader
		reader := strings.NewReader(string(b))
		req.Body = ioutil.NopCloser(reader)
	}

}
