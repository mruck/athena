package sql

import (
	"fmt"

	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
	"github.com/xwb1989/sqlparser"
)

// parseNode searches for a parameter value.  If found, it allocates a tainted query
// and populates the param and column fields
func parseNode(node sqlparser.SQLNode, param string) (*TaintedQuery, error) {
	switch node := node.(type) {
	case *sqlparser.SQLVal:
		if string(node.Val) != param {
			return nil, nil
		}
		// Found it
		return &TaintedQuery{Param: param}, nil
	case *sqlparser.ComparisonExpr:
		switch node.Operator {
		case sqlparser.InStr:
			fallthrough
		case sqlparser.EqualStr:
			fallthrough
		case sqlparser.GreaterThanStr:
			// Check val for a match
			match, err := parseNode(node.Right, param)
			if err != nil {
				return nil, err
			}
			if match == nil {
				return nil, nil
			}
			col := node.Left.(*sqlparser.ColName).Name.String()
			match.Column = col
			return match, nil
		default:
			err := fmt.Errorf("unhandled operator: %v", node.Operator)
			return nil, errors.WithStack(err)
		}
	case sqlparser.ValTuple:
		node = []sqlparser.Expr(node)
		for _, expr := range node {
			match, err := parseNode(expr, param)
			if err != nil {
				return nil, err
			}
			if match != nil {
				return match, nil
			}
		}
		return nil, nil
	case *sqlparser.Where:
		// Sanity checking on where
		if node.Type != "where" {
			log.Fatalf("where.Type == %v\n", node.Type)
		}
		return parseNode(node.Expr, param)
	case *sqlparser.Select:
		return parseSelect(node, param)
	case *sqlparser.Update:
		return parseUpdate(node, param)
	case *sqlparser.Delete:
		return parseDelete(node, param)
	case *sqlparser.Insert:
		return parseInsert(node, param)
	case *sqlparser.Subquery:
		return parseNode(node.Select, param)
	case *sqlparser.DDL:
		return parseDDL(node, param)
	}
	err := fmt.Errorf("searching for params %q and hit unhandled node type: %T", param, node)
	util.PrettyPrintStruct(node)
	return nil, errors.WithStack(err)
}

func parseDDL(ddl *sqlparser.DDL, param string) (*TaintedQuery, error) {
	// The parameter is neither present in the old table name nor in the new
	// table name
	if ddl.Table.Name.String() != param && ddl.NewName.Name.String() != param {
		return nil, nil
	}
	// TODO: should also check if param name is equivalent to column names
	// being created

	// Found it
	var action Action
	switch ddl.Action {
	case sqlparser.CreateStr:
		action = Create
	case sqlparser.DropStr:
		action = Drop
	case sqlparser.TruncateStr:
		action = Truncate
	case sqlparser.AlterStr:
		action = Alter
	default:
		err := fmt.Errorf("failed to match ddl.Action = %s", ddl.Action)
		log.Error(errors.WithStack(err))
		return nil, err
	}
	query := &TaintedQuery{Param: param, Table: param, Action: action}
	return query, nil
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
	match, err := parseNode(stmt.Where, param)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, nil
	}

	// Found the param.  Check if the table name has been set, if it has then
	// this belongs to a nested query and a different table.
	if match.Table != "" {
		return match, nil
	}

	// If not, then this is the tainty query so set the table name
	name, err := parseTableName(stmt.From)
	if err != nil {
		return nil, err
	}
	match.Table = name

	match.Action = Select

	return match, nil
}

func parseUpdate(stmt *sqlparser.Update, param string) (*TaintedQuery, error) {
	match, err := parseNode(stmt.Where, param)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, nil
	}

	// Found the param.  Check if the table name has been set, if it has then
	// this belongs to a nested query and a different table.
	if match.Table != "" {
		return match, nil
	}

	// Parse table name
	name, err := parseTableName(stmt.TableExprs)
	if err != nil {
		return nil, err
	}
	match.Table = name

	match.Action = Update
	return match, nil
}

func parseDelete(stmt *sqlparser.Delete, param string) (*TaintedQuery, error) {
	match, err := parseNode(stmt.Where, param)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, nil
	}

	// Found the param.  Check if the table name has been set, if it has then
	// this belongs to a nested query and a different table.
	if match.Table != "" {
		return match, nil
	}

	// Parse table name
	name, err := parseTableName(stmt.TableExprs)
	if err != nil {
		return nil, err
	}
	match.Table = name

	match.Action = Delete

	return match, nil
}

func parseQuery(query string, param string) (*TaintedQuery, error) {
	stmt, err := sqlparser.Parse(query)
	//util.PrettyPrintStruct(stmt)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	//util.PrettyPrintStruct(stmt)
	return parseNode(stmt, param)
}
