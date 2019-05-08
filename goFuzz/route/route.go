package route

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mruck/athena/goFuzz/util"
	"github.com/pkg/errors"
)

// Route object contains metadata about a route
type Route struct {
	Path            string
	Method          string
	QueryParams     string
	DynamicSegments []string
	BodyParams      string
}

// JSONRoute is a Route object in JSON formm as dumped by rails
type JSONRoute struct {
	Path     string
	Verb     string
	Segments []string
}

// toRoute converts a JSONRoute object to a Route object
func (jsonified *JSONRoute) toRoute() *Route {
	return &Route{Path: jsonified.Path, Method: jsonified.Verb,
		DynamicSegments: jsonified.Segments}
}

// RoutesPath is a path to routes dumped by Rails
// TODO: get this info from rails by sending a request
const RoutesPath = "tests/routes.json"

// LoadRoutes reads routes from shared mount and loads them into memory
func LoadRoutes() []*Route {
	// Unmarshal into a JSON struct
	JSONRoutes := []JSONRoute{}
	util.MustUnmarshalFile(RoutesPath, JSONRoutes)

	// Convert the list of JSONRoute structs to a list of Route objects
	routes := make([]*Route, len(JSONRoutes))
	for _, JSONroute := range JSONRoutes {
		routes = append(routes, JSONroute.toRoute())
	}
	return routes
}

// ToHTTPRequest converts a route to an http.Request
func (route *Route) ToHTTPRequest() *http.Request {
	url := fmt.Sprintf("http://overwriteMe.com%s", route.Path)
	req, err := http.NewRequest(url, route.Method, nil)
	if err != nil {
		err = errors.Wrap(err, "")
		log.Fatalf("%+v\n", err)
	}
	return req
}
