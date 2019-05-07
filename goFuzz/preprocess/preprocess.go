package preprocess

import (
	"net/http"

	"github.com/pkg/errors"
)

// GetLogin parses a har file with login information and returns
// a series of GO requests to replicate that behavior
func GetLogin(harPath string) ([]*http.Request, error) {
	har, err := unmarshalHar(harPath)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return har.toRequests()
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
