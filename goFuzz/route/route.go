package route

import (
	"regexp"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/har"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/param"
	"github.com/mruck/athena/goFuzz/sql/sqlparser"
	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
)

const path = "path"

// SiblingMethod contains mutation state for sibling methods
// on the same path
type SiblingMethod struct {
	Method string
	Param  *[]*param.Param
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
	Params         []*param.Param
	SiblingMethods *[]*SiblingMethod
	// Har entries for this route
	Entries *[]har.Entry
}

// New initializes parameter state, stores it in the sibling method list,
// then allocates a route with this information
func New(path string, method string, meta *spec.Operation, siblingMethods *[]*SiblingMethod) *Route {
	// Initialize object to keep track of state for each param
	params := param.InitializeParamState(meta.Parameters)

	// Discourse swagger didn't enumerate all path params, so manually check and add if necessary
	param.CheckForPathParams(path, &params)

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
		Params: params, Re: re, Entries: entries}
}

// UpdateQueries maps each tainted query to a parameter
func (route *Route) UpdateQueries(queries []sqlparser.TaintedQuery) bool {
	newQueries := false
	for i, query := range queries {
		for _, param := range route.Params {
			for _, metadata := range param.GetMetadata() {
				// Found a match
				if metadata.Values[0] == query.Param {
					// This is the first time seeing this query, we got new coverage
					if metadata.TaintedQuery == nil {
						newQueries = true
						metadata.TaintedQuery = &queries[i]
					}
				}
			}
		}
	}
	return newQueries
}

// CurrentParams stringifies the most recent params sent and returns them as a list
func (route *Route) CurrentParams() []string {
	params := []string{}
	for _, param := range route.Params {
		// We never set this parameter
		if param.Next == nil {
			continue
		}
		latest := param.LatestValues()
		params = append(params, latest...)
	}
	return params
}

// hasPathParams checks whether or not a route has path params
func (route *Route) hasPathParams() bool {
	for _, param := range route.Params {
		if param.In == "path" {
			return true
		}
	}
	return false
}

// PrettyPrint most recent request sent at log level specified or level `info`
// if level is nil
func (route *Route) PrettyPrint(logFn log.Fn) {
	if logFn == nil {
		logFn = log.Infof
	}

	if route.hasPathParams() {
		// Print the canonicalized path i.e. /about/{type}.json
		logFn("%s %s", route.Method, route.Path)
	}

	// Get the most recent request sent
	req, err := route.ToHTTPRequest()
	if err != nil {
		log.Warn(err)
		return
	}

	// Pretty print request that was sest
	httpclient.PrettyPrintRequest(req, logFn)
}

// LogError logs an error with the context of the most recent request sent
func (route *Route) LogError(traceback error) {

	// Print the context
	route.PrettyPrint(log.Errorf)

	// Log original error
	log.Error(traceback)
}

// Testing only: generate dummy data
func (route *Route) MockData() {
	for _, param := range route.Params {
		param.MockData()
	}
}
