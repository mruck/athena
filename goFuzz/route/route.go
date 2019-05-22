package route

import (
	"regexp"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/har"
	"github.com/mruck/athena/goFuzz/param"
	"github.com/mruck/athena/lib/util"
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
	// Har entries for this route
	Entries *[]har.Entry
}

// New initializes parameter state, stores it in the sibling method list,
// then allocates a route with this information
func New(path string, method string, meta *spec.Operation, siblingMethods *[]*SiblingMethod) *Route {
	// Initialize object to keep track of state for each param
	state := param.InitializeParamState(meta.Parameters)

	// Allocate an object to keep track of har entries for the route
	entries := &[]har.Entry{}

	// Create a regex for the path so we can match against Har requests
	// i.e. /t/9 from the har should be matched against /t/{id}
	re, err := canonicalizePath(path)
	// TODO: should continue if this fails
	util.Must(err == nil, "%+v\n", err)

	// TODO: implement this!
	// Update the sibling meta data so it contains this method's
	// mutation state
	//sibling := &SiblingMethod{Method: method, State: &state}
	//*siblingMethods = append(*siblingMethods, sibling)

	return &Route{Path: path, Method: method, Meta: meta,
		State: state, Re: re, Entries: entries}
}

// Mutate mutates parameters in a route, setting param.State.Next
// for each parameter, or nil if the paramater shouldn't be sent
func (route *Route) Mutate() {
	for _, param := range route.State {
		param.Mutate()
	}
}
