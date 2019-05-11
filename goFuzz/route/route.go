package route

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/param"
	"github.com/mruck/athena/goFuzz/util"
	"github.com/pkg/errors"
)

// SiblingMethod contains mutation state for sibling methods
// on the same path
type SiblingMethod struct {
	Method string
	State  *[]*param.State
}

// Route contains static information from a swagger, and dynamic mutation state
type Route struct {
	// key in Spec.Swagger.Paths.Paths
	Path string
	// field in Spec.Swagger.Path.Paths[Path].Get/Put/etc
	Method string
	// regexp is for matching path with a har request
	Re *regexp.Regexp
	// value for field in Spec.Swagger.Path.Path[Path].Get
	Meta *spec.Operation
	// Mutation state for each parameter object
	State          []*param.State
	SiblingMethods *[]*SiblingMethod
}

// New initializes parameter state, stores it in the sibling method list,
// then allocates a route with this information
func New(path string, method string, meta *spec.Operation, siblingMethods *[]*SiblingMethod) *Route {
	// Initialize object to keep track of state for each param
	state := []*param.State{}
	// Create a regex for the path so we can match against Har requests
	// i.e. /t/9 from the har should be matched against /t/{id}
	re, err := canonicalizePath(path)
	// TODO: should continue if this fails
	util.Must(err == nil, "%+v\n", err)
	// Update the sibling meta data so it contains this method's
	// mutation state
	sibling := &SiblingMethod{Method: method, State: &state}
	*siblingMethods = append(*siblingMethods, sibling)
	return &Route{Path: path, Method: method, Meta: meta,
		State: state, SiblingMethods: siblingMethods, Re: re}
}

// canonicalizePath creates a regexp for the path so it can be matched
// against a har request
func canonicalizePath(path string) (*regexp.Regexp, error) {
	re := regexp.MustCompile(`/\{[^/]+\}`)
	pathRegexp := re.ReplaceAllString(path, "/([^/]+)")
	re, err := regexp.Compile(pathRegexp)
	util.Must(err == nil, "%+v\n", errors.WithStack(err))
	return re, err
}

// FindRouteByPath searches for a route with matching path and method
func FindRouteByPath(routes []*Route, path string, method string) *Route {
	for _, route := range routes {
		if route.Method == method && route.Re.Match([]byte(path)) {
			return route
		}
	}
	return nil
}

// ReadSwagger file into memory
func ReadSwagger(path string) *spec.Swagger {
	swagger := &spec.Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}

// Check if a route is blacklisted
func blacklisted(path string) bool {
	return strings.Contains(path, "readonly") ||
		strings.Contains(path, "logout")
}

// FromSwagger loads routes from swagger file
func FromSwagger(path string) []*Route {
	swagger := ReadSwagger(path)
	// All routes
	routes := []*Route{}
	for path, operations := range swagger.Paths.Paths {
		// This path is blacklisted
		if ok := blacklisted(path); ok {
			continue
		}

		// All methods on the same path should share an object pointing
		// to parameter state
		siblingMethods := &[]*SiblingMethod{}
		if operations.Get != nil {
			route := New(path, "GET", operations.Get, siblingMethods)
			routes = append(routes, route)
		}
		if operations.Put != nil {
			route := New(path, "PUT", operations.Put, siblingMethods)
			routes = append(routes, route)
		}
		if operations.Post != nil {
			route := New(path, "POST", operations.Post, siblingMethods)
			routes = append(routes, route)
		}
		if operations.Delete != nil {
			route := New(path, "DELETE", operations.Delete, siblingMethods)
			routes = append(routes, route)
		}
		if operations.Options != nil {
			route := New(path, "OPTIONS", operations.Options, siblingMethods)
			routes = append(routes, route)
		}
		if operations.Head != nil {
			route := New(path, "HEAD", operations.Head, siblingMethods)
			routes = append(routes, route)
		}
		if operations.Patch != nil {
			route := New(path, "PATCH", operations.Patch, siblingMethods)
			routes = append(routes, route)
		}
	}

	return routes
}

// ToHTTPRequest converts a route to an http.Request
func (route *Route) ToHTTPRequest() (*http.Request, error) {
	return nil, nil
	//url := fmt.Sprintf("http://overwriteMe.com%s", route.Path)
	//req, err := http.NewRequest(route.Method, url, nil)
	//if err != nil {
	//	return nil, errors.Wrap(err, "")
	//}
	//return req, nil
}
