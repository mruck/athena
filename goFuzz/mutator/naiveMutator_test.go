package mutator

import (
	"testing"

	"github.com/mruck/athena/goFuzz/swagger"
	"github.com/mruck/athena/lib/util"
	"github.com/stretchr/testify/require"
)

// PetStoreExpanded is path to pet store swagger with refs expanded (for testing)
const PetStoreExpanded = "../tests/petstore_expanded.json"

func TestPathParam(t *testing.T) {
	// Get the param obj
	path := "/pet/{petId}"
	method := "get"
	paramName := "petId"
	param, err := swagger.MockParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	// Embed a metadata obj
	swagger.EmbedParam(param)

	// Mutate the leaf nodes
	mutateParam(param)

	util.PrettyPrintStruct(param, nil)
	// Check we actually embedded something
	//meta := ReadOneMetadata(param)
	//require.Equal(t, []interface{}{}, meta.Values)

}
