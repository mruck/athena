package preprocess

import (
	"net/http"

	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/util"
)

// GetLogin parses a har file with login information and returns
// a series of GO requests to replicate that behavior
func GetLogin(harPath string) ([]*http.Request, error) {
	har := unmarshalHar(harPath)
	return har.toRequests()
}

// GetCorpus parses a harfile, initializing relevant data in
// the list of routes.  It returns the har requests as an ordered list of route.Routes
func GetCorpus(routes []*route.Route, harPath string) []*route.Route {
	har := unmarshalHar(harPath)
	corpus, err := har.InitializeRoutes(routes)
	util.Must(err == nil, "%+v\n", err)
	return corpus
}
