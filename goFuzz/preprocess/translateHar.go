package preprocess

import (
	"fmt"
	"net/url"

	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/util"
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
		// TODO: eventually log this and move in on
		util.Must(err == nil, "%+v\n", errors.WithStack(err))
		route := route.FindRouteByPath(routes, url.Path, entry.Request.Method)
		// We didn't find this route in the swagger spec
		if route == nil {
			fmt.Printf("Skipping: %v %v\n", entry.Request.Method, url.Path)
			continue
		}
		// Initialize Har data inside route
		initializeRoute(route, entry)
		corpus = append(corpus, route)
	}
	return corpus, nil
}
