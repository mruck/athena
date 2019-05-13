package swagger

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const PetStore = "tests/petstore.json"
const PetStoreExpanded = "tests/petstore_expanded.json"

// TODO: make this a table test
func petStoreGenerate(path string, method string) (map[string]interface{}, error) {
	return Generate(PetStoreExpanded, path, method)
}

//// TestPathParam tests a path parameter that has no schema
//func TestPathParam(t *testing.T) {
//	path := "/pet/{petId}"
//	method := "get"
//	obj, err := Generate(PetStoreExpanded, path, method)
//	require.NoError(t, err)
//	_, ok := obj["petId"]
//	require.True(t, ok)
//}
//
//// TestObj tests generating a random object for body parameters
//func TestStruct(t *testing.T) {
//	path := "/store/order"
//	method := "post"
//	obj, err := Generate(PetStoreExpanded, path, method)
//	require.NoError(t, err)
//	_, ok := obj["body"]
//	require.True(t, ok)
//}
//
func TestArray(t *testing.T) {
	path := "/pet/findByStatus"
	method := "get"
	obj, err := petStoreGenerate(path, method)
	require.NoError(t, err)
	_, ok := obj["status"]
	require.True(t, ok)
}

// TestExpandedSchema makes refs are getting expanded properly
func TestExpandSchema(t *testing.T) {
	// Check that refs are present
	ref := "$ref"
	refFile, err := ioutil.ReadFile(PetStore)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(refFile), ref))

	// Expand
	tmpfile, err := ioutil.TempFile("/tmp", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	err = Expand(PetStore, tmpfile.Name())
	require.NoError(t, err)
	expanded, err := ioutil.ReadFile(tmpfile.Name())
	require.NoError(t, err)

	// Check that refs are expanded
	require.False(t, strings.Contains(string(expanded), ref))
}
