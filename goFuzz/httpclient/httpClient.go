package httpclient

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/pkg/errors"
)

func prerunHook(client *http.Client, requests []*http.Request) error {
	for _, request := range requests {
		resp, err := client.Do(request)
		if err != nil {
			return errors.Wrap(err, "")
		}
		fmt.Printf("status code: %v\n", resp.StatusCode)
		if resp.StatusCode == 403 {
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
