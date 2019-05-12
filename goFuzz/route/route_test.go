package route

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadSwagger(t *testing.T) {
	swagger := ReadSwagger("../tests/dummySwagger.json")
	//util.PrettyPrintStruct(swagger)
	// Check that a field is correct
	description := swagger.Paths.Paths["/categories.json"].Post.Description
	require.Equal(t, "Create a new category", description)
}

// Just make sure we don't hit an unmarshaling error
func TestReadDiscourseSwagger(t *testing.T) {
	_ = ReadSwagger("../tests/discourseSwagger.json")
	//util.PrettyPrintStruct(swagger)
}

func TestFromSwagger(t *testing.T) {
	routes := FromSwagger("../tests/dummySwagger.json")
	// Check a random field
	require.Equal(t, routes[0].Method, "GET")
}

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
