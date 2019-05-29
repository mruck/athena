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

// parseNode searches for a parameter value.  If found, it allocates a tainted query
// and populates the param and column fields
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
		log.Fatalf("where.Type == %v\n", where.Type)
	}
	return parseNode(where.Expr, param)
}

func parseTableName(exprs sqlparser.TableExprs) (string, error) {
	if len(exprs) != 1 {
		log.Fatal("there was more than 1 table expresion\n")
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
		log.Fatal("Match is nil!\n")
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

func parseUpdate(stmt *sqlparser.Update, param string) (*TaintedQuery, error) {
	match, err := parseWhere(stmt.Where, param)
	if err != nil {
		return nil, err
	}
	if match == nil {
		// We should only call parseQuery when we know the param is present in the string
		log.Fatal("Match is nil!\n")
	}
	match.CRUD = Update

	// Parse table name
	name, err := parseTableName(stmt.TableExprs)
	if err != nil {
		return nil, err
	}
	match.Table = name

	return match, nil
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
		//util.PrettyPrintStruct(stmt)
		return parseUpdate(stmt, param)
	case *sqlparser.Delete:
		util.PrettyPrintStruct(stmt)
	default:
		log.Fatalf("Unhandled statement type: %T\n", stmt)
	}
	return nil, nil
}
