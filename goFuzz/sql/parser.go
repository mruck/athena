package sql

import (
	"fmt"

	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
	"github.com/xwb1989/sqlparser"
)

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

func parseDelete(stmt *sqlparser.Delete, param string) (*TaintedQuery, error) {
	match, err := parseWhere(stmt.Where, param)
	if err != nil {
		return nil, err
	}
	if match == nil {
		// We should only call parseQuery when we know the param is present in the string
		log.Fatal("Match is nil!\n")
	}
	match.CRUD = Delete

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
		return parseUpdate(stmt, param)
	case *sqlparser.Delete:
		return parseDelete(stmt, param)
	default:
		log.Fatalf("Unhandled statement type: %T\n", stmt)
	}
	return nil, nil
}
