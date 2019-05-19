package httpclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/mruck/athena/goFuzz/util"
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
		return false, errors.Wrap(err, "")
	}

	for i := 0; i < maxAttempts; i++ {
		fmt.Printf("Polling %v\n", url)
		resp, err := cli.Do(request)
		if err != nil {
			return false, errors.Wrap(err, "")
		}
		// Target app is up
		// TODO: the 404 status code is wonky.  Change rails fork endpoint to
		// return 200
		if resp.StatusCode == 404 {
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
	req.Host = cli.URL.Host
	req.URL.Host = cli.URL.Host
	resp, err := cli.Client.Do(req)
	cli.updateStatusCodes(resp.StatusCode)

	return resp, err
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

// PrettyPrintRequest pretty prints http request
func PrettyPrintRequest(req *http.Request) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	url := req.URL.Path
	if req.Method == "GET" {
		if req.URL.RawQuery != "" {
			url += "?" + req.URL.RawQuery
		}
	}
	log.Printf("%v %v\n", req.Method, url)
	if req.Body != nil {
		b, err := ioutil.ReadAll(req.Body)
		util.Must(err == nil, "%+v\n", errors.WithStack(err))
		dst := map[string]interface{}{}
		err = json.Unmarshal(b, &dst)
		util.Must(err == nil, "%+v\n", errors.WithStack(err))
		util.PrettyPrintStruct(dst)
		// Reinitialize the reader
		reader := strings.NewReader(string(b))
		req.Body = ioutil.NopCloser(reader)
	}

}
