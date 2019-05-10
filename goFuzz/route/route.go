package route

import (
	"net/http"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/param"
	"github.com/mruck/athena/goFuzz/util"
)

// SiblingMethod contains mutation state for sibling methods
// on the same path
type SiblingMethod struct {
	Method string
	State  *param.State
}

// Route contains static information from a swagger, and dynamic mutation state
type Route struct {
	// key in Spec.Swagger.Paths.Paths
	Path string
	// field in Spec.Swagger.Path.Paths[Path].Get/Put/etc
	Method string
	// value for field in Spec.Swagger.Path.Path[Path].Get
	Meta           *spec.Operation
	State          *param.State
	SiblingMethods *[]*SiblingMethod
}

// New initializes parameter state, stores it in the sibling method list,
// then allocates a route with this information
func New(path string, method string, meta *spec.Operation, siblingMethods *[]*SiblingMethod) *Route {
	// Initialize object to keep track of state
	state := &param.State{}
	// Update the sibling meta data so it contains this methods
	// mutation state
	sibling := &SiblingMethod{Method: method, State: state}
	*siblingMethods = append(*siblingMethods, sibling)
	return &Route{Path: path, Method: method, Meta: meta,
		State: state, SiblingMethods: siblingMethods}
}

// ReadSwagger file into memory
func ReadSwagger(path string) *spec.Swagger {
	swagger := &spec.Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}

// LoadRoutes from swagger file
func LoadRoutes(path string) ([]*Route, error) {
	swagger := ReadSwagger(path)
	// All routes
	routes := []*Route{}
	for path, operations := range swagger.Paths.Paths {
		// All methods on the same path should share an object pointing
		// to parameter state
		siblingMethods := &[]*SiblingMethod{}
		if operations.Get != nil {
			// Create a route
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

	return routes, nil
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
