package mutator

import (
	"testing"

	"github.com/mruck/athena/goFuzz/swagger"
	"github.com/mruck/athena/lib/util"
	"github.com/stretchr/testify/require"
)

// PetStoreExpanded is path to pet store swagger with refs expanded (for testing)
const PetStoreExpanded = "../tests/petstore_expanded.json"

// Test mutating a path param
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
	// TODO: Check we actually added a value with the same Not Nil check
	// as below
}

// Test mutating body
func TestBody(t *testing.T) {
	// Get our param
	path := "/store/order"
	method := "post"
	paramName := "body"
	param, err := swagger.MockParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	// Embed a metadata obj
	swagger.EmbedParam(param)

	// Mutate the leaf nodes
	mutateParam(param)

	// Check that metadata.Values for each leaf node has
	// valid data from our mutation
	metadata := swagger.ReadAllMetadata(param)
	for _, meta := range metadata {
		require.NotNil(t, meta.Values)
	}
}

func TestArrayWithPrimativeMetadata(t *testing.T) {
}

func TestMetaArrayWithObj(t *testing.T) {
}
