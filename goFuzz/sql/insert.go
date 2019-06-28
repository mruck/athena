package sql

import (
	"fmt"

	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
	"github.com/xwb1989/sqlparser"
)

// parseRow searches each val in a row for a parameter. If it's found, return
// the index into the row
func parseRow(exprs sqlparser.Exprs, param string) int {
	for i, expr := range exprs {
		switch node := expr.(type) {
		case *sqlparser.SQLVal:
			if string(node.Val) == param {
				return i
			}
		// TODO: unimplemented
		case *sqlparser.OrExpr:
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
	err := fmt.Errorf("failed to find param in list of inserted values")
	return -1, errors.WithStack(err)
}

// Iterate through columns and return the column name at the given index.  This allows
// us to map a value inserted to a column
func iterateColumns(index int, columns sqlparser.Columns) (string, error) {
	if len(columns) == 0 {
		err := "unhandled: insert with no columns provided"
		log.Fatal(err)
	}
	for i, column := range columns {
		if i == index {
			return column.String(), nil
		}
	}
	err := fmt.Errorf("failed to find index %v in columns", index)
	log.Fatal(err)
	return "", errors.WithStack(err)
}

func parseInsert(stmt *sqlparser.Insert, param string) (*TaintedQuery, error) {
	index, err := parseRows(stmt.Rows, param)
	if err != nil {
		return nil, err
	}
	// TODO: columns will not always be present! If not then I need to connect
	// to db to see what cols are then cache
	column, err := iterateColumns(index, stmt.Columns)
	if err != nil {
		return nil, err
	}
	table := stmt.Table.Name.String()
	tainted := &TaintedQuery{
		Column: column,
		Param:  param,
		Table:  table,
		Action: Insert,
	}
	return tainted, nil
}
