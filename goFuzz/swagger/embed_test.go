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
	embedParam(param)

	// Check we actually embedded something
	values := readValues(param)
	require.Equal(t, []interface{}{}, values)
}

// Test by iterating through all routes
func TestFromRoutes(t *testing.T) {

}

// Test from our athena defined parameter obj
func TestFromParam(t *testing.T) {

}

// Test emedding for a body with an object
func TestBodyObj(t *testing.T) {
	// Get our param
	path := "/store/order"
	method := "post"
	paramName := "body"
	param, err := getParam(PetStoreExpanded, path, method, paramName)
	require.NoError(t, err)

	embedParam(param)

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

	extensions := param.Schema.Properties["status"].VendorExtensible.Extensions
	metadata := extensions[xreferential].(*metadata)
	val := metadata.Values[0].(string)
	// Make sure the leaf points to the new value
	require.Equal(t, "hello", val)
}

//// Test setting metadata for primitive array
//func TestArrayWithPrimativeMetadata(t *testing.T) {
//	// Get a parameter
//	path := "/pet/findByStatus"
//	method := "get"
//	paramName := "status"
//	param, err := getParam(PetStoreExpanded, path, method, paramName)
//	require.NoError(t, err)
//	// Val is the newly generated value
//	val := GenerateAny(param)
//	// Stored is the stored newly generated value
//	stored := readNewestValue(&param.Items.VendorExtensible)
//	// Make sure the recently generated value is equal to the most recently
//	// stored value
//	require.Equal(t, val, stored)
//}
