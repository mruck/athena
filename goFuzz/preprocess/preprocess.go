package preprocess

import (
	"net/http"

	"github.com/mruck/athena/goFuzz/har"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/lib/util"
)

// GetLogin parses a har file with login information and returns
// a series of GO requests to replicate that behavior
func GetLogin(harPath string) ([]*http.Request, error) {
	harObj := har.UnmarshalHar(harPath)
	return harObj.ToHTTPRequests()
}

// GetCorpus parses a harfile, initializing relevant data in
// the list of routes.  It returns the har requests as an ordered list of route.Routes
func GetCorpus(routes []*route.Route, harPath string) []*route.Route {
	// Read in our corpus
	harData := har.UnmarshalHar(harPath)
	// Initialize route objects from the har
	corpus, err := InitializeRoutes(routes, harData)
	util.Must(err == nil, "%+v\n", err)
	return corpus
}
