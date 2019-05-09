package mutator

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mruck/athena/goFuzz/coverage"
	"github.com/mruck/athena/goFuzz/route"
)

// Mutate specifies required functions to be defined on a mutator class
type Mutate interface {
}

// Mutator contains state for mutating
type Mutator struct {
	Routes     []*route.Route
	routeIndex int
	Coverage   *coverage.Coverage
}

func seedMutator(corpus []*http.Request) {
}

const coveragePath = "/tmp/results/coverage.json"

// New creates a new mutator
func New(corpus []*http.Request, routes []*route.Route) *Mutator {
	// TODO: use the corpus to seed the mutator.  It will probs also change
	// the type of mutation alg we pick?
	seedMutator(corpus)
	return &Mutator{Routes: routes, routeIndex: -1, Coverage: coverage.New(coveragePath)}
}

// Mutate pick the next route and mutates the parameters
func (mutator *Mutator) Mutate() *route.Route {
	mutator.routeIndex++
	// We've exhausted all routes
	if mutator.routeIndex >= len(mutator.Routes) {
		return nil
	}
	return mutator.Routes[mutator.routeIndex]
}

// Next picks the route, mutates the parameters, and formats it as a request
func (mutator *Mutator) Next() *http.Request {
	route := mutator.Mutate()
	// We are done
	if route == nil {
		return nil
	}
	req, err := route.ToHTTPRequest()
	if err != nil {
		// TODO: this route failed. Log to a file and mutate again
		log.Fatalf("%+v\n", err)
	}
	return req
}

// UpdateCoverage parses the response and updates source code, parameter and
// query coverage
func (mutator *Mutator) UpdateCoverage(resp *http.Response) error {
	err := mutator.Coverage.Update()
	fmt.Printf("Delta: %v\n", mutator.Coverage.Delta)
	fmt.Printf("Cumulative: %v\n", mutator.Coverage.Cumulative)
	if err != nil {
		return err
	}
	return nil
}
