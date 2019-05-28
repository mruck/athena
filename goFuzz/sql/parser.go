package sql

import (
	"encoding/base64"
	"strings"

	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
	"github.com/xwb1989/sqlparser"
)

// Sql operations
const (
	Update = iota
	Select
	Delete
	Insert
)

// Analyze searches for parameters in the given queries
func Analyze(params []string, queries []string) ([]TaintedQuery, error) {
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

// TaintedQuery is a sql query tainted with user controlled data
type TaintedQuery struct {
	Param string
	// Raw query to run to get comparable results
	Query  string
	Table  string
	Column string
	// CRUD action
	Method string
}

func parseNode(node sqlparser.SQLNode, param string) (*TaintedQuery, error) {
	//log.Infof("Type: %T\n", node)
	switch stmt := node.(type) {
	// Leaf
	case *sqlparser.ComparisonExpr:
		// Check val for a match
		sqlval := stmt.Right.(*sqlparser.SQLVal)
		if string(sqlval.Val) != param {
			return nil, nil
		}
		// Found it
		col := stmt.Left.(*sqlparser.ColName)
		match := &TaintedQuery{Param: param, Column: col.Name.String()}
		return match, nil
	}
	return nil, nil
}

func parseWhere(where *sqlparser.Where, param string) (*TaintedQuery, error) {
	if where.Type != "where" {
		log.Fatalf("where.Type == %v\n", where.Type)
	}
	return parseNode(where.Expr, param)
}

func parseQuery(query string, param string) (*TaintedQuery, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		util.PrettyPrintStruct(stmt)
		return parseWhere(stmt.Where, param)
		// If we matched a param, parse the FROM clause to identify table name
		// Add method
	case *sqlparser.Insert:
		// Cast to a list of values
		values := stmt.Rows.(sqlparser.Values)
		log.Infof("type: %T\n", values[0])
		util.PrettyPrintStruct(values[0])
		log.Infof("type: %T\n", values[0][0])
		util.PrettyPrintStruct(values[0][0])
		sqlVal := values[0][0].(*sqlparser.SQLVal)
		data, _ := base64.StdEncoding.DecodeString(string(sqlVal.Val))
		log.Infof("Decoding %v as %v\n", string(sqlVal.Val), string(data))
	case *sqlparser.Update:
		util.PrettyPrintStruct(stmt)
	case *sqlparser.Delete:
		util.PrettyPrintStruct(stmt)
	default:
		log.Panicf("Unhandled statement type: %T\n", stmt)
	}
	return nil, nil
}
