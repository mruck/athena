package mutator

import (
	"net/http"

	"github.com/mruck/athena/goFuzz/route"
)

// Mutate specifies required functions to be defined on a mutator class
type Mutate interface {
}

// Mutator contains state for mutating
type Mutator struct {
	Routes []*route.Route
}

func seedMutator(corpus []*http.Request) {
}

// New creates a new mutator
func New(corpus []*http.Request, routes []*route.Route) *Mutator {
	// TODO: use the corpus to seed the mutator.  It will probs also change
	// the type of mutation alg we pick?
	seedMutator(corpus)
	return &Mutator{Routes: routes}
}

// Mutate pick the next route and mutates the parameters
func (mutator *Mutator) Mutate() *route.Route {
	return nil
}

// Next picks the route, mutates the parameters, and formats it as a request
func (mutator *Mutator) Next() *http.Request {
	route := mutator.Mutate()
	request := route.ToHTTPRequest()
	return request
}

// UpdateCoverage parses the response and updates source code, parameter and
// query coverage
func (mutator *Mutator) UpdateCoverage(resp *http.Response) {
}
