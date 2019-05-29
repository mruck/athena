package sql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//func TestAnalyze(t *testing.T) {
//	params := []string{"sunnyvale", "los altos", "marin"}
//	queries := []string{
//		"statement: insert into cities (name, temp) values ('sunnyvale', 60);",
//		"statement: insert into cities (name, temp) values ('san jose', 67);",
//	}
//	parser := Parser{"parser.py"}
//	matches, err := parser.Analyze(params, queries)
//	require.NoError(t, err)
//	fmt.Println(matches)
//	//require.NotNil(t, matches)
//}

func TestSelect(t *testing.T) {
	sql := "SELECT * FROM mytable WHERE city = 'sunnyvale';"
	match, err := parseQuery(sql, "sunnyvale")
	require.NoError(t, err)
	require.Equal(t, "city", match.Column)
	require.Equal(t, "mytable", match.Table)
}

func TestInsertOneRow(t *testing.T) {
	sql := "insert into cities (name, temp) values ('san jose', 67);"
	// TODO: test searching for non stringified params? or should i stringify them before to make easier?
	match, err := parseQuery(sql, "san jose")
	require.NoError(t, err)
	require.Equal(t, "city", match.Column)
	require.Equal(t, "mytable", match.Table)
}

func TestInsertManyRows(t *testing.T) {
}
