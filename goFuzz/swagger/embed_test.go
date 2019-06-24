package swagger

import (
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
			if *leaf == correctData {
				found = true
			}
		}
		require.True(t, found)
	}

	//require.Equal(t, leaves, data)
	//util.PrettyPrintStruct(param, nil)

	// Check that the metadata leaves stored in the top level are correct

	// Check that the self referential pointers are updated when we modify
	// the metada leaves

	//// Check the entire body stored vs generated
	//stored := readNewestValue(&param.Schema.VendorExtensible)
	//require.Equal(t, val, stored)

	//// Check one of the stored leaf nodes
	//schema := param.Schema.Properties["complete"]
	//val = readNewestValue(&schema.VendorExtensible)
	//require.NotNil(t, val)
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
