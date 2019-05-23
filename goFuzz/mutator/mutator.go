package mutator

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mruck/athena/goFuzz/coverage"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/lib/database"
	"github.com/mruck/athena/lib/exception"
	"github.com/mruck/athena/lib/util"
)

// Mutate specifies required functions to be defined on a mutator class
type Mutate interface {
}

// Mutator contains state for mutating
type Mutator struct {
	Routes            []*route.Route
	routeIndex        int
	Coverage          *coverage.Coverage
	ExceptionsManager *exception.ExceptionsManager
	TargetID          string
}

// New creates a new mutator
func New(routes []*route.Route, corpus []*route.Route) *Mutator {
	// Connect to mongodb to log exceptions
	db := database.MustGetDatabase(database.MongoDbPort, "athena")
	manager := exception.NewExceptionsManager(db, exception.Path)

	// Make the order deterministic for debugging.  Order routes alphabetically
	route.Order(routes)

	// TODO: do something with the corpus
	return &Mutator{
		Routes:            routes,
		routeIndex:        -1,
		Coverage:          coverage.New(coverage.Path),
		ExceptionsManager: manager,
		TargetID:          util.MustGetTargetID(),
	}
}

func (mutator *Mutator) specialRoute() *route.Route {
	routeEnvVar := os.Getenv("ROUTE")
	if routeEnvVar == "" {
		return nil
	}
	for _, route := range mutator.Routes {
		if route.Path == routeEnvVar {
			return route
		}
	}
	log.Println("ROUTE env var set but didn't find")
	return nil
}

// Mutate picks the next route and mutates the parameters
func (mutator *Mutator) Mutate() *route.Route {
	// User specified route
	if route := mutator.specialRoute(); route != nil {
		return route
	}

	// We didn't get new coverage, next route
	if mutator.Coverage.Delta == 0 {
		mutator.routeIndex++
		// We've exhausted all routes
		if mutator.routeIndex >= len(mutator.Routes) {
			return nil
		}
	}
	route := mutator.Routes[mutator.routeIndex]
	// Mutate each parameter
	route.Mutate()
	return route
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

func (mutator *Mutator) currentRoute() *route.Route {
	return mutator.Routes[mutator.routeIndex]
}

// UpdateState parses the response and updates source code, parameter and
// query coverage
func (mutator *Mutator) UpdateState(resp *http.Response) error {
	err := mutator.Coverage.Update()
	fmt.Printf("Delta: %v\n", mutator.Coverage.Delta)
	fmt.Printf("Cumulative: %v\n", mutator.Coverage.Cumulative)
	if err != nil {
		return err
	}
	route := mutator.currentRoute()
	return mutator.ExceptionsManager.Update(route.Path, route.Method, mutator.TargetID)
}
