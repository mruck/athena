package sql

import (
	"fmt"
	"strings"

	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
	"github.com/xwb1989/sqlparser"
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

func parseNode(node sqlparser.SQLNode, param string) (*TaintedQuery, error) {
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
	// Handle in
	// Handle and/or
	return nil, nil
}

func parseWhere(where *sqlparser.Where, param string) (*TaintedQuery, error) {
	if where.Type != "where" {
		log.Panicf("where.Type == %v\n", where.Type)
	}
	return parseNode(where.Expr, param)
}

func parseTableName(exprs sqlparser.TableExprs) (string, error) {
	if len(exprs) != 1 {
		log.Panic("there was more than 1 table expresion\n")
		return "", fmt.Errorf("there was more than 1 table expression")
	}
	aliasedTableExpr := exprs[0].(*sqlparser.AliasedTableExpr)
	tableName := aliasedTableExpr.Expr.(sqlparser.TableName)
	return tableName.Name.String(), nil
	//log.Infof("Type == %T\n", aliasedTableExpr.Expr)
}

func parseSelect(stmt *sqlparser.Select, param string) (*TaintedQuery, error) {
	match, err := parseWhere(stmt.Where, param)
	if err != nil {
		return nil, err
	}
	if match == nil {
		// We should only call parseQuery when we know the param is present in the string
		log.Panic("Match is nil!\n")
	}

	match.CRUD = Select

	// Parse table name
	name, err := parseTableName(stmt.From)
	if err != nil {
		return nil, err
	}
	match.Table = name

	return match, nil
}

// parseRow searches each val in a row for a parameter. If it's found, return
// the index into the row
func parseRow(exprs sqlparser.Exprs, param string) int {
	for i, expr := range exprs {
		sqlVal := expr.(*sqlparser.SQLVal)
		if string(sqlVal.Val) == param {
			return i
		}
	}
	return -1
}

// This query inserts multiple rows, search each row for our param
func parseRows(insertRows sqlparser.InsertRows, param string) (int, error) {
	rows := insertRows.(sqlparser.Values)
	for _, row := range rows {
		row := sqlparser.Exprs(row)
		index := parseRow(row, param)
		// Found it
		if index >= 0 {
			return index, nil
		}
	}
	err := fmt.Errorf("failed to find param in list of values")
	log.Panic(err)
	return -1, errors.WithStack(err)
}

func parseInsert(stmt *sqlparser.Insert, param string) (*TaintedQuery, error) {
	util.PrettyPrintStruct(stmt)
	// Items are inserted as list.  Figure out the index of our parameter
	//values := stmt.Rows.(sqlparser.Values)
	index, err := parseRows(stmt.Rows, param)
	log.Infof("Index: %v", index)
	if err != nil {
		return nil, err
	}
	// Map that index to a list of columns
	// Identify the table

	//log.Infof("type: %T\n", values[0])
	//util.PrettyPrintStruct(values[0])
	//log.Infof("type: %T\n", values[0][0])
	//util.PrettyPrintStruct(values[0][0])
	//sqlVal := values[0][0].(*sqlparser.SQLVal)
	//data, _ := base64.StdEncoding.DecodeString(string(sqlVal.Val))
	//log.Infof("Decoding %v as %v\n", string(sqlVal.Val), string(data))
	return nil, nil
}

// How to handle generic values like `1`, etc?
func parseQuery(query string, param string) (*TaintedQuery, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		return parseSelect(stmt, param)
	case *sqlparser.Insert:
		return parseInsert(stmt, param)
	case *sqlparser.Update:
		util.PrettyPrintStruct(stmt)
	case *sqlparser.Delete:
		util.PrettyPrintStruct(stmt)
	default:
		log.Panicf("Unhandled statement type: %T\n", stmt)
	}
	return nil, nil
}
