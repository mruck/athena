package route

import (
	"testing"

	"github.com/mruck/athena/goFuzz/swagger"
	"github.com/mruck/athena/lib/util"
	"github.com/stretchr/testify/require"
)

type regexpTest struct {
	input  string
	output string
	err    error
}

func TestRegExp(t *testing.T) {
	table := []regexpTest{
		{"/a/b/{p1}/c", "/a/b/([^/]+)/c", nil},
		{"/d/{p1}/e/{p2}", "/d/([^/]+)/e/([^/]+)", nil},
		{"/d/{p1}/e/f", "/d/([^/]+)/e/f", nil},
		{"/a/b", "/a/b", nil},
		{"/a/b/c", "/a/b/c", nil},
		{"/", "/", nil},
	}
	for _, test := range table {
		output, err := canonicalizePath(test.input)
		if test.err != nil {
			require.Error(t, err)
		}
		require.Equal(t, test.output, output.String())
	}
}

func TestFromSwagger(t *testing.T) {
	routes := FromSwagger("../tests/dummySwagger.json")
	// Check a random field
	require.Equal(t, routes[0].Method, "GET")
}

func TestEmdbed(t *testing.T) {
	routes := FromSwagger("../tests/discourseSwagger.json")
	for _, route := range routes {
		for _, param := range route.Params {
			swagger.EmbedParam(&param.Parameter)
		}
	}
	util.PrettyPrintStruct(routes, nil)
}
