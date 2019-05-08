package preprocess

import (
	"io"
	"net/http"
	"strings"

	"github.com/mruck/athena/goFuzz/util"
	"github.com/pkg/errors"
)

type header struct {
	Name  string
	Value string
}

type cookie struct {
	Name     string
	Value    string
	Expires  string
	HTTPOnly bool
	Secure   bool
}

type param struct {
	Name  string
	Value string
}

type postData struct {
	MimeType string
	Text     string
	Params   []param
}

type query struct {
	Name  string
	Value string
}

// Request key in har
type request struct {
	Method      string
	URL         string
	Headers     []header
	QueryString []query
	Cookies     []cookie
	PostData    postData
}

// Response key in har
type response struct {
	Status int
}

// Entry in har file
type entry struct {
	Request  request
	Response response
}

// Log in har file
type log struct {
	Entries []entry
}

// Har json representation
type Har struct {
	Log log
}

// toHTTPRequest converts a har request to a http.Request
func (req *request) toHTTPRequest() (*http.Request, error) {
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

// toRequest converts a har struct to a list of http.Requests
func (har *Har) toRequests() ([]*http.Request, error) {
	entries := har.Log.Entries
	requests := make([]*http.Request, len(entries))
	for i, entry := range entries {
		// Convert each Har request to http.Request
		req, err := entry.Request.toHTTPRequest()
		if err != nil {
			return nil, err
		}
		requests[i] = req
	}
	return requests, nil
}

// unmarshalHar() takes in a har file and returns a Har struct
func unmarshalHar(harPath string) *Har {
	har := &Har{}
	util.MustUnmarshalFile(harPath, har)
	return har
}
