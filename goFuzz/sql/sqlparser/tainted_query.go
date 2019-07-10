package sqlparser

import (
	"fmt"
	"regexp"

	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
)

// Action on database
type Action int

// CRUD Sql operations
const (
	Update Action = iota
	Select
	Delete
	Insert
	Create
	Truncate
	Drop
	Alter
	Rename
)

// TaintedQuery is a sql query tainted with user controlled data
type TaintedQuery struct {
	// Parameter value that we searched for and identified inside the query
	Param string
	// Raw query to run to get comparable results
	Query  string
	Table  string
	Column string
	Action Action
}

// Determine whether or not param is present inside query string by using
// a regex rather than strings.Contains which returns lots of false postives
func matchParam(param string, query string) bool {
	expr := fmt.Sprintf("= '%s' ", param)
	matched, err := regexp.Match(expr, []byte(query))
	if err != nil {
		log.Warnf("%+v", errors.WithStack(err))
		return false
	}
	return matched
}
