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

// TestPathParam tests a path parameter that has no schema
func TestPathParam(t *testing.T) {
	path := "/pet/{petId}"
	method := "get"
	obj, err := Generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	_, ok := obj["petId"]
	require.True(t, ok)
}

//// TestObj tests generating a random object
//func TestObj(t *testing.T) {
//	path := "/store/order"
//	method := "post"
//	obj, err := Generate(PetStoreExpanded, path, method)
//	require.NoError(t, err)
//	fmt.Printf("%v\n", obj)
//}

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
