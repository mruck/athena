package mutator

import (
	"net/http"
	"os"
	"strings"

	"github.com/mruck/athena/goFuzz/coverage"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/sql/postgres"
	"github.com/mruck/athena/goFuzz/sql/sqlparser"
	"github.com/mruck/athena/lib/database"
	"github.com/mruck/athena/lib/exception"
	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
)

// Mutator contains state for mutating
type Mutator struct {
	SQLParser  *sqlparser.Parser
	Routes     []*route.Route
	routeIndex int
	// Source code coverage
	SrcCoverage *coverage.Coverage
	// Did we get new query coverage?
	QueryDelta        bool
	ExceptionsManager *exception.ExceptionsManager
	TargetID          string
	DBLog             *postgres.PGLog
	// user specified route via env vars ROUTE and METHOD
	userRoute *route.Route
}

// New creates a new mutator
func New(routes []*route.Route, corpus []*route.Route) *Mutator {
	// Connect to mongodb to log exceptions
	db := database.MustGetDatabase(database.MongoDbPort, "athena")
	manager := exception.NewExceptionsManager(db, exception.Path)

	// Make the order deterministic for debugging.  Order routes alphabetically
	route.Order(routes)

	// Get a new pg log and seek to the end
	pgLog := postgres.NewLog()
	pgLog.Seek()

	mutator := &Mutator{
		Routes:            routes,
		routeIndex:        -1,
		SrcCoverage:       coverage.New(coverage.Path),
		ExceptionsManager: manager,
		TargetID:          util.MustGetTargetID(),
		DBLog:             pgLog,
		SQLParser:         sqlparser.NewParser(),
	}

	// Check if user specified route, and if so update our mutator to reflect that
	mutator.getUserRoute()

	return mutator
}

// get user specified route
func (mutator *Mutator) getUserRoute() {
	routeEnvVar := os.Getenv("ROUTE")
	if routeEnvVar == "" {
		return
	}
	method := os.Getenv("METHOD")
	if method == "" {
		return
	}

	for i, route := range mutator.Routes {
		if strings.EqualFold(route.Path, routeEnvVar) {
			if strings.EqualFold(route.Method, method) {
				// On mutator we increment the index, so start at -1
				mutator.routeIndex = i - 1
				mutator.userRoute = route
				return
			}
		}
	}
	log.Infof("ROUTE=%s METHOD=%s but couldn't find a match", routeEnvVar, method)
	os.Exit(1)
}

func (mutator *Mutator) exitImmediately() {
	//  This is our first time calling mutator
	if mutator.SrcCoverage.Delta == 0 {
		return
	}
	// We've mutated once and want to exit now
	if os.Getenv("EXIT") == "1" {
		os.Exit(0)
	}
}

// Mutate picks the next route and mutates the parameters
func (mutator *Mutator) Mutate() *route.Route {
	// Exit after 1 request
	mutator.exitImmediately()

	// We didn't get new coverage, next route
	if mutator.SrcCoverage.Delta == 0 && !mutator.QueryDelta {
		mutator.routeIndex++
		// A user specified route was provided
		if mutator.userRoute != nil {
			// We are done mutating the user specified route so we are done here
			if mutator.Routes[mutator.routeIndex] != mutator.userRoute {
				return nil
			}
		}
		// We've exhausted all routes
		if mutator.routeIndex >= len(mutator.Routes) {
			return nil
		}
	}
	route := mutator.Routes[mutator.routeIndex]

	// Mutate each parameter
	mutator.MutateRoute(route)

	return route
}

// Next picks the route, mutates the parameters, and formats it as a request
func (mutator *Mutator) Next() *http.Request {
	// Ask the mutator for the next route
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

func (mutator *Mutator) logStats(route *route.Route) {
	route.PrettyPrint(nil)
	log.Infof("Delta: %v", mutator.SrcCoverage.Delta)
	log.Infof("Cumulative: %v", mutator.SrcCoverage.Cumulative)
	mutator.SQLParser.PrettyPrint()
}

// UpdateState parses the response and updates source code, parameter and
// query coverage
func (mutator *Mutator) UpdateState(resp *http.Response) error {
	// Get current route
	route := mutator.currentRoute()

	// Update source code coverage
	err := mutator.SrcCoverage.Update()
	if err != nil {
		return err
	}

	// Read log dumped by postgres
	queries, err := mutator.DBLog.Next()
	if err != nil {
		return err
	}

	// Triage postgres log for errors, hints, etc
	mutator.DBLog.Triage()

	// Search for params present in queries
	params := route.CurrentParams()
	taintedQueries, err := mutator.SQLParser.Search(queries, params)
	if err != nil {
		return err
	}

	// Log various stats above cov, queries, etc
	mutator.logStats(route)

	// Update route with queries...new queries count as coverage
	// TODO: queries will need to be canonicalized. For now lets just
	// compare tainted queries cause we can just use the struct to compare
	mutator.QueryDelta = route.UpdateQueries(taintedQueries)

	// Check for sql inj
	sqlparser.CheckForSQLInj(queries, params)

	// Store any new exceptions
	return mutator.ExceptionsManager.Update(route.Path, route.Method, mutator.TargetID)
}

// LogError logs an error with context from the most recent request sent
func (mutator *Mutator) LogError(err error) {
	// Get current route
	route := mutator.currentRoute()
	route.LogError(err)
}
