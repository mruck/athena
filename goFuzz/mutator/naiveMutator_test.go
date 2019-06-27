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

	// Format the data
	val := swagger.Format(param)
	require.NotNil(t, val)
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
	// Format the data
	val := swagger.Format(param)

	// Check that we got back a map with all necessary keys
	dict, ok := val.(map[string]interface{})
	require.True(t, ok)
	_, ok = dict["complete"]
	require.True(t, ok)
	_, ok = dict["id"]
	require.True(t, ok)
	_, ok = dict["petId"]
	require.True(t, ok)
	_, ok = dict["quantity"]
	require.True(t, ok)
	_, ok = dict["shipDate"]
	require.True(t, ok)
	_, ok = dict["status"]
	require.True(t, ok)
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
	// Format the data
	val := swagger.Format(param)

	// Check that we were given a string array
	array, ok := val.([]interface{})
	require.True(t, ok)
	_, ok = array[0].(string)
	require.True(t, ok)
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

	// Format the data
	val := swagger.Format(param)

	// Check that we were given an array of objects (specifically
	// each object is a map[string]interface{}
	array, ok := val.([]interface{})
	require.True(t, ok)

	_, ok = array[0].(map[string]interface{})
	require.True(t, ok)
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
	// Format the data
	val := swagger.Format(param)

	// Check our nested dict
	dict, ok := val.(map[string]interface{})
	require.True(t, ok)

	// Extract the nested dict
	nested, ok := dict["category"]
	require.True(t, ok)

	dict2, ok := nested.(map[string]interface{})
	require.True(t, ok)

	// Check nested keys are present
	_, ok = dict2["id"]
	require.True(t, ok)
	_, ok = dict2["name"]
	require.True(t, ok)
}
