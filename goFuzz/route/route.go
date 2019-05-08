package route

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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

// RoutesPath is a path to a file of routes dumped by rails.
// The file contains a list of new line separated JSONRoute
// objects.
// TODO: get this info from rails by sending a request
const RoutesPath = "/tmp/results/routes.json"

// LoadRoutes reads routes from shared mount and loads them into memory
func LoadRoutes() []*Route {
	// Parse new line separated routes file
	data, err := ioutil.ReadFile(RoutesPath)
	if err != nil {
		err := errors.Wrap(err, "")
		log.Fatalf("%+v\n", err)
	}
	routeStrings := strings.Split(string(data), "\n")
	// The file ends with a blank new line so trim
	// TODO: clean this up from the rails side
	routeStrings = routeStrings[:len(routeStrings)-1]

	// Unmarshal into a JSON struct
	JSONroutes := make([]*JSONRoute, len(routeStrings))
	for i, routeString := range routeStrings {
		route := JSONRoute{}
		if err := json.Unmarshal([]byte(routeString), &route); err != nil {
			err = errors.Wrap(err, "")
			log.Fatalf("%+v\n", err)
		}
		JSONroutes[i] = &route
	}

	// Initialize the universal route data structure
	// TODO: is it necessary to have the JSON metadata struct above?
	routes := make([]*Route, len(JSONroutes))
	for i, JSONRoute := range JSONroutes {
		routes[i] = JSONRoute.toRoute()
	}
	return routes
}

// ToHTTPRequest converts a route to an http.Request
func (route *Route) ToHTTPRequest() (*http.Request, error) {
	url := fmt.Sprintf("http://overwriteMe.com%s", route.Path)
	req, err := http.NewRequest(url, route.Method, nil)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return req, nil
}
