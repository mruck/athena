package preprocess

import (
	"net/http"
)

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
