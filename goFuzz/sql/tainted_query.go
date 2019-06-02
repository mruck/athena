package sql

import (
	"strings"

	"github.com/mruck/athena/lib/util"
)

// CRUD action on db
type CRUD int

// CRUD Sql operations
const (
	Update CRUD = iota
	Select
	Delete
	Insert
)

// TaintedQuery is a sql query tainted with user controlled data
type TaintedQuery struct {
	Param string
	// Raw query to run to get comparable results
	Query  string
	Table  string
	Column string
	CRUD   CRUD
}

// search searches for parameters in the given queries
func search(params []string, queries []string) ([]TaintedQuery, error) {
	// params or queries are empty, return
	if len(params) == 0 || len(queries) == 0 {
		return nil, nil
	}
	matches := []TaintedQuery{}
	for _, query := range queries {
		// Search this query for each param
		for _, param := range params {
			// Do a simple string check before searching
			if !strings.Contains(query, param) {
				continue
			}
			match, err := parseQuery(param, query)
			util.Must(err == nil, "%+v\n", err)
			if match != nil {
				matches = append(matches, *match)
			}
		}
	}
	return matches, nil
}
