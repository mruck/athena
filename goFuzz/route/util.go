package route

// Util functions on the Route object

import (
	"regexp"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/util"
	"github.com/pkg/errors"
)

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
