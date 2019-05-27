package sql

import (
	"testing"

	"github.com/xwb1989/sqlparser/dependency/sqltypes"
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

func TestLib(t *testing.T) {
	for i := 0; i < 256; i++ {
		sqltypes.SQLEncodeMap[i] = byte(i)
		sqltypes.SQLDecodeMap[i] = byte(i)
	}

	//sql := "SELECT * FROM mytable WHERE a = 'abc';"
	sql := "insert into cities (name, temp) values ('san jose', 67);"
	_ = ParseQuery(sql)
}
