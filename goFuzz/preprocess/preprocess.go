package preprocess

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// Request key in har
type request struct {
	Method      string
	URL         string
	Headers     []header
	QueryString []string // ?
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

// fileToHar() takes in a har file and returns a Har struct
func fileToHar(harPath string) (*Har, error) {
	data, err := ioutil.ReadFile(harPath)
	if err != nil {
		return nil, err
	}
	har := &Har{}
	err = json.Unmarshal(data, har)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v", har)
	return har, nil
}

// harToRequest converts a har struct to a []http.Request
func harToRequest(har *Har) ([]*http.Request, error) {
	return nil, nil
}

// GetLogin parses a har file with login information and returns
// a series of GO requests to replicate that behavior
func GetLogin() []*http.Request {
	return nil
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
