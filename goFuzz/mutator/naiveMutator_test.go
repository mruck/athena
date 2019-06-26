package mutator

import (
	"testing"

	"github.com/mruck/athena/goFuzz/swagger"
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

	// Check that metadata.Values for each leaf node has
	// valid data from our mutation
	metadata := swagger.ReadAllMetadata(param)
	for _, meta := range metadata {
		require.NotNil(t, meta.Values)
	}
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
	path := "/pet/findByStatus"
	method := "get"
	paramName := "status"
	param, err := swagger.MockParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	// Embed our param
	swagger.EmbedParam(param)

	// Mutate the leaf nodes
	mutateParam(param)

	// Check that metadata.Values for each leaf node has
	// valid data from our mutation
	metadata := swagger.ReadAllMetadata(param)
	for _, meta := range metadata {
		require.NotNil(t, meta.Values)
	}
	//util.PrettyPrintStruct(param, nil)
}

func TestMetaArrayWithObj(t *testing.T) {
	path := "/user/createWithArray"
	method := "post"
	paramName := "body"
	param, err := swagger.MockParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	// Embed our param
	swagger.EmbedParam(param)

	// Mutate the leaf nodes
	mutateParam(param)

	// Check that metadata.Values for each leaf node has
	// valid data from our mutation
	metadata := swagger.ReadAllMetadata(param)
	for _, meta := range metadata {
		require.NotNil(t, meta.Values)
	}
	//util.PrettyPrintStruct(param, nil)
}

func TestNestedDict(t *testing.T) {
	// Get our param
	path := "/pet"
	method := "put"
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
