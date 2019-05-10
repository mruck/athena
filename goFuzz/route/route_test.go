package route

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadSwagger(t *testing.T) {
	swagger := ReadSwagger("dummySwagger.json")
	//util.PrettyPrintStruct(swagger)
	// Check that a field is correct
	description := swagger.Paths.Paths["/categories.json"].Post.Description
	require.Equal(t, "Create a new category", description)
}

// Just make sure we don't hit an unmarshaling error
func TestReadDiscourseSwagger(t *testing.T) {
	_ = ReadSwagger("discourseSwagger.json")
	//util.PrettyPrintStruct(swagger)
}

func TestFromSwagger(t *testing.T) {
	routes := FromSwagger("dummySwagger.json")
	// Check a random field
	require.Equal(t, routes[0].Method, "GET")
}
