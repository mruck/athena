package swagger

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
