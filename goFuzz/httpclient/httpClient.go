package httpclient

import (
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/pkg/errors"
)

const maxAttempts = 60

// Try every 2 seconds
const interval = 2
const healthCheckRoute = "/rails/info/pluralization"

// HealthCheck checks if a hard coded rails fork endpoint is up
func HealthCheck(url string) (bool, error) {
	client := &http.Client{}
	url += healthCheckRoute
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, errors.Wrap(err, "")
	}
	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Do(request)
		if err != nil {
			return false, errors.Wrap(err, "")
		}
		// Target app is up
		if resp.StatusCode == 200 {
			return true, nil
		}
		time.Sleep(time.Second * interval)
	}

	// We never got a heartbeat
	return false, nil
}

func prerunHook(client *http.Client, requests []*http.Request) error {
	for _, request := range requests {
		_, err := client.Do(request)
		if err != nil {
			return errors.Wrap(err, "")
		}
	}
	return nil
}

// newClient allocates an http client with a cookie jar
func newClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &http.Client{Jar: jar}, nil
}

// New initializes a HTTPClient using the state provided
func New(requests []*http.Request) (*http.Client, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}
	// Allocate http client
	err = prerunHook(client, requests)
	return client, err
}
