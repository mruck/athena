package sql

import (
	"encoding/base64"

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

//// Analyze searches for parameters in the given queries
//func Analyze(params []string, queries []string) ([]ParamQuery, error) {
//	// params or queries are empty, return
//	if len(params) == 0 || len(queries) == 0 {
//		return nil, nil
//	}
//	for _, query := range queries {
//		AnalyzeQuery(params, query)
//	}
//	return nil, nil
//}

// ParamQuery data matching a parameter to a query
type ParamQuery struct {
	Param string
	// Raw query to run to get comparable results
	Query  string
	Table  string
	Column string
	// CRUD action
	Method string
}

func ParseNode(node sqlparser.SQLNode) {
	log.Infof("Type: %T\n", node)
	switch stmt := node.(type) {
	case *sqlparser.ComparisonExpr:
		// Todo: store the operator?
		// Only hand = for now to map to table/col, add other stuff
		// later
		ParseNode(stmt.Left)
		ParseNode(stmt.Right)
	case *sqlparser.ColName:
		log.Infof("col name: %v", stmt.Name)
	case *sqlparser.SQLVal:
		log.Infof("col val: %v", string(stmt.Val))
	}

}
func ParseWhere(where *sqlparser.Where) {
	if where.Type != "where" {
		log.Fatalf("where.Type == %v\n", where.Type)
	}
	ParseNode(where.Expr)
}

func ParseQuery(query string) error {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return errors.WithStack(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		util.PrettyPrintStruct(stmt)
		ParseWhere(stmt.Where)
		// If we matched a param, parse the FROM clause to identify table name
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
	return nil
}
