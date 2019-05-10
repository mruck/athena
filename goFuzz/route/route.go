package route

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/mruck/athena/goFuzz/param"
	"github.com/pkg/errors"
)

// Route object contains metadata about a route
type Route struct {
	Path            string
	Method          string
	QueryParams     []*param.Param
	DynamicSegments []*param.Param
	BodyParams      []*param.Param
}

// JSONRoute is a Route object in JSON formm as dumped by rails
type JSONRoute struct {
	Path     string
	Verb     string
	Segments []string
}

// Mutate params on route
func (route *Route) Mutate() {
	//	if route.DynamicSegments != nil {
	//		for _, param := range route.DynamicSegments {
	//			param.Mutate()
	//		}
	//	}
}

// toRoute converts a JSONRoute object to a Route object
func (jsonified *JSONRoute) toRoute() *Route {
	segments := make([]*param.Param, len(jsonified.Segments))
	for i, segment := range jsonified.Segments {
		segments[i] = param.New(segment)
	}

	return &Route{Path: jsonified.Path, Method: jsonified.Verb,
		QueryParams: nil, BodyParams: nil,
		DynamicSegments: segments}
}

// RoutesPath is a path to a file of routes dumped by rails.
// The file contains a list of new line separated JSONRoute
// objects.
// TODO: get this info from rails by sending a request
const RoutesPath = "/tmp/results/routes.json"

// LoadRoutes reads routes from shared mount and loads them into memory
func LoadRoutes() []*Route {
	// TODO: ignore blacklisted route, ie RO route
	// Parse new line separated routes file
	data, err := ioutil.ReadFile(RoutesPath)
	if err != nil {
		err := errors.Wrap(err, "")
		log.Fatalf("%+v\n", err)
	}
	routeStrings := strings.Split(string(data), "\n")
	// The file ends with a blank new line so trim
	// TODO: change rails to dump straight json so i can trivially load
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
	req, err := http.NewRequest(route.Method, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return req, nil
}
