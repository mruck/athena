package swagger

import (
	"reflect"
	"testing"

	"github.com/mruck/athena/lib/util"
	"github.com/stretchr/testify/require"
)

// Ensure we embed metadata correctly for primitive parameters.
// In this case the primitive parameter is a path parameter.
func TestPrimitive(t *testing.T) {
	// Get the param obj
	path := "/pet/{petId}"
	method := "get"
	paramName := "petId"
	param, err := getParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	// Embed
	EmbedParam(param)

	// Check we actually embedded something
	meta := readOneMetadata(param)
	require.Equal(t, []interface{}{}, meta.Values)
}

// Test emedding for a body with an object
func TestBodyObj(t *testing.T) {
	// Get our param
	path := "/store/order"
	method := "post"
	paramName := "body"
	param, err := getParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	EmbedParam(param)

	// Read what the leaves should contain
	data := []metadata{}
	util.MustUnmarshalFile("test/body.metadata", &data)

	// Read the leaves we collected
	leaves := readMetadata(param)

	// Compare the 2
	for _, leaf := range leaves {
		found := false
		for _, correctData := range data {
			if reflect.DeepEqual(*leaf, correctData) {
				found = true
			}
		}
		require.True(t, found)
	}

	// Check that the self referential pointers are updated when we modify
	// the metadata leaves
	for _, leaf := range leaves {
		// Set an arbitrary value
		if leaf.Schema.Description != "" {
			leaf.Values = append(leaf.Values, "hello")
		}
	}

	// Read the leaf value
	extensions := param.Schema.Properties["status"].VendorExtensible.Extensions
	metadata := extensions[xreferential].(*metadata)
	val := metadata.Values[0].(string)

	// Make sure the leaf points to the new value
	require.Equal(t, "hello", val)
}

// Test setting metadata for primitive array for query param
func TestArrayWithPrimativeMetadata(t *testing.T) {
	// Get a parameter
	path := "/pet/findByStatus"
	method := "get"
	paramName := "status"
	param, err := getParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	// Embed our param
	EmbedParam(param)

	// Check that our metadata is present.  This is a query parameter, so it's
	// single level so we don't need to embed the schema
	// Check we actually embedded something
	meta := readOneMetadata(param)
	require.Equal(t, []interface{}{}, meta.Values)
}

// Test storing metadata for an array storing complex objs.
// This can only be present in a body param
func TestMetaArrayWithObj(t *testing.T) {
	// Get a parameter
	path := "/user/createWithArray"
	method := "post"
	paramName := "body"
	param, err := getParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	// Embed our param
	EmbedParam(param)

	// Read what the leaves should contain
	data := []metadata{}
	util.MustUnmarshalFile("test/body_array.metadata", &data)

	// Read the leaves we collected
	leaves := readMetadata(param)

	// Compare the 2
	for _, leaf := range leaves {
		found := false
		for _, correctData := range data {
			if reflect.DeepEqual(*leaf, correctData) {
				found = true
			}
		}
		require.True(t, found)
	}

}
