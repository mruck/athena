package preprocess

import (
	"net/url"

	"github.com/mruck/athena/goFuzz/route"
	"github.com/pkg/errors"
)

// initializeRoute takes a har entry and initializes the associated route object
func initializeRoute(route *route.Route, harEntry entry) {
}

// InitializeRoutes initializes a list of routes given information
// from the provided har.  Returns an ordered list of routes that
// reflects the har file.
func (har *Har) InitializeRoutes(routes []*route.Route) ([]*route.Route, error) {
	corpus := []*route.Route{}
	for _, entry := range har.Log.Entries {
		url, err := url.Parse(entry.Request.URL)
		if err != nil {
			// TODO: if we hit this case we should log it and continue
			return nil, errors.WithStack(err)
		}
		route := route.FindRouteByPath(routes, url.Path, entry.Request.Method)
		if route == nil {
			// TODO: if we hit this case we should log it and continue
			return nil, errors.WithStack(err)
		}
		// Initialize Har data inside route
		initializeRoute(route, entry)
		corpus = append(corpus, route)
	}
	return corpus, nil
}
