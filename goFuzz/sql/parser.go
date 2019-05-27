package sql

import (
	"encoding/base64"
	"fmt"
	"log"

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

func ParseQuery(query string) error {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return errors.WithStack(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		util.PrettyPrintStruct(stmt.Where)
	case *sqlparser.Insert:
		// Cast to a list of values
		values := stmt.Rows.(sqlparser.Values)
		log.Printf("type: %T\n", values[0])
		util.PrettyPrintStruct(values[0])
		log.Printf("type: %T\n", values[0][0])
		util.PrettyPrintStruct(values[0][0])
		sqlVal := values[0][0].(*sqlparser.SQLVal)
		data, _ := base64.StdEncoding.DecodeString(string(sqlVal.Val))
		fmt.Printf("Decoding %v as %v\n", string(sqlVal.Val), string(data))

	case *sqlparser.Update:
		util.PrettyPrintStruct(stmt)
	case *sqlparser.Delete:
		util.PrettyPrintStruct(stmt)
	default:
		log.Panicf("Unhandled statement type: %T\n", stmt)
	}
	return nil
}
