package preprocess

import (
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func newReq(req request) (*http.Request, error) {
	body := io.Reader(nil)
	// This isn't a GET request, check for a body
	if req.Method != "GET" {
		body = strings.NewReader(req.PostData.Text)
	}
	newReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	// Update headers
	for _, header := range req.Headers {
		newReq.Header.Add(header.Name, header.Value)
	}
	return newReq, nil
}

// harToRequest converts a har struct to a []http.Request
func harToRequest(har *Har) ([]*http.Request, error) {
	entries := har.Log.Entries
	requests := make([]*http.Request, len(entries))
	for i, entry := range entries {
		// Allocate new request
		req, err := newReq(entry.Request)
		if err != nil {
			return nil, err
		}
		requests[i] = req
	}
	return requests, nil
}

// TODO: this should be in the shared mount.  Not sure a way around hard
// coding this
const harPath = "tests/login_har.json"

// GetLogin parses a har file with login information and returns
// a series of GO requests to replicate that behavior
func GetLogin() ([]*http.Request, error) {
	har, err := unmarshalHar(harPath)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return harToRequest(har)
}

// Corpus contains Go formated requests to use as initial corpus
type Corpus struct {
	Requests []http.Request
}

// GetCorpus parses Har file, formating relevant info like url, headers, params,
// etc and formating into a list of requests
func GetCorpus() *Corpus {
	return nil
}
