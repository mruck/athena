package har

import (
	"io"
	"net/http"
	"strings"

	"github.com/mruck/athena/goFuzz/util"
	"github.com/pkg/errors"
)

// Header field from har file
type Header struct {
	Name  string
	Value string
}

// Cookie field from har file
type Cookie struct {
	Name     string
	Value    string
	Expires  string
	HTTPOnly bool
	Secure   bool
}

// Param field from har file
type Param struct {
	Name  string
	Value string
}

// PostData field from har file
type PostData struct {
	MimeType string
	Text     string
	Params   []Param
}

// Query field from har file
type Query struct {
	Name  string
	Value string
}

// Request field in har
type Request struct {
	Method      string
	URL         string
	Headers     []Header
	QueryString []Query
	Cookies     []Cookie
	PostData    PostData
}

// Response field in har
type Response struct {
	Status int
}

// Entry in har file
type Entry struct {
	Request  Request
	Response Response
}

// Log in har file
type Log struct {
	Entries []Entry
}

// Har json representation
type Har struct {
	Log Log
}

// toHTTPRequest converts a har request to a http.Request
func (req *Request) toHTTPRequest() (*http.Request, error) {
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

// ToHTTPRequests converts a har struct to a list of http.Requests
func (har *Har) ToHTTPRequests() ([]*http.Request, error) {
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

// UnmarshalHar takes in a har file and returns a Har struct
func UnmarshalHar(harPath string) *Har {
	har := &Har{}
	util.MustUnmarshalFile(harPath, har)
	return har
}
