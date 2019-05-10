package preprocess

import (
	"fmt"

	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/util"
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
		// Canonicalize har path
		path, err := entry.Request.canonicalizePath()
		// TODO: in the future, log this and continue
		util.Must(err == nil, "%+v\n", err)
		route := route.FindRouteByPath(routes, path, entry.Request.Method)
		// We didn't find this route in the swagger spec
		if route == nil {
			fmt.Printf("Skipping: %v %v\n", entry.Request.Method, path)
			continue
		}
		// Initialize Har data inside route
		initializeRoute(route, entry)
		corpus = append(corpus, route)
	}
	return corpus, nil
}
