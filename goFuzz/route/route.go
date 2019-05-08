package route

import (
	"net/http"

	"github.com/mruck/athena/goFuzz/util"
)

// Route object contains metadata about a route
type Route struct {
	Path            string
	Method          string
	QueryParams     string
	DynamicSegments string
	BodyParams      string
}

// JSONRoute is a Route object in JSON formm as dumped by rails
type JSONRoute struct {
	Path     string
	Verb     string
	Segments []string
}

// RoutesPath is a path to routes dumped by Rails
// TODO: get this info from rails by sending a request
const RoutesPath = "tests/routes.json"

// LoadRoutes reads routes from shared mount and loads them into memory
func LoadRoutes() []*Route {
	JSONRoutes := []JSONRoute{}
	util.MustUnmarshalFile(RoutesPath, JSONRoutes)

	return nil
}

// ToHTTPRequest converts a route to an http.Request
func (route *Route) ToHTTPRequest() *http.Request {
	return nil
}
