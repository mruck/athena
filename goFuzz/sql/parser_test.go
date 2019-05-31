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
	match, err := parseQuery(sql, "san jose")
	require.NoError(t, err)
	require.Equal(t, "name", match.Column)
	require.Equal(t, "cities", match.Table)
}

// TODO:
// Test insertion with no columns specified (see if we fail gracefully)
func TestInsertNoCol(t *testing.T) {
	return
	sql := "insert into cities values ('san jose', 67);"
	match, err := parseQuery(sql, "san jose")
	require.NoError(t, err)
	require.Equal(t, "name", match.Column)
	require.Equal(t, "cities", match.Table)
}

// TODO: test searching for non stringified params? or should i stringify
// them before to make easier? i.e. with values ('san jose', 67) if i'm looking
// for 67 should param be 67 or '67'. either stringify both sql insertion
// vals and param or make sure param is native type
func TestTypes(t *testing.T) {
}

func TestInsertManyRows(t *testing.T) {
	sql := "insert into cities (name, temp) values ('san jose', 67), ('sunnyvale', 60), " +
		"('palo alto', 58);"
	match, err := parseQuery(sql, "sunnyvale")
	require.NoError(t, err)
	require.Equal(t, "name", match.Column)
	require.Equal(t, "cities", match.Table)
}

// Test updating a single column
func TestUpdate(t *testing.T) {
	sql := "update cities set temp = 30 where name = 'sunnyvale'"
	match, _ := parseQuery(sql, "sunnyvale")
	require.Equal(t, "name", match.Column)
	require.Equal(t, "cities", match.Table)
}

// Test updating multiple columns
func TestUpdateMultipleCol(t *testing.T) {
	sql := "update cities set temp = 30, state = 'california'  where name = 'sunnyvale'"
	match, _ := parseQuery(sql, "sunnyvale")
	require.Equal(t, "name", match.Column)
	require.Equal(t, "cities", match.Table)
}

// Test a greater than operator to make sure we handle, even if it's not exactly
// how we want to handle
func TestGreaterThan(t *testing.T) {
	sql := "update cities set temp = 30, state = 'california'  where name > 'sunnyvale'"
	match, _ := parseQuery(sql, "sunnyvale")
	require.Equal(t, "name", match.Column)
	require.Equal(t, "cities", match.Table)
}

func TestDelete(t *testing.T) {
	sql := "delete from cities where name = 'sunnyvale'"
	match, _ := parseQuery(sql, "sunnyvale")
	require.Equal(t, "name", match.Column)
	require.Equal(t, "cities", match.Table)
	//util.PrettyPrintStruct(match)
}

// Test update statement with nested select
func TestUpdateFromSelect(t *testing.T) {

}

func TestIn2(t *testing.T) {
	sql := "select name, temp from cities where temp in (50, 60);"
	_, _ = parseQuery(sql, "sunnyvale")
}
