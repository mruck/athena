package swagger

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/util"
	"github.com/stretchr/testify/require"
)

const PetStoreExpanded = "test/petstore_expanded.json"
const PetStore = "test/petstore.json"

// TestPathParam tests a path parameter
func TestPathParam(t *testing.T) {
	path := "/pet/{petId}"
	method := "get"
	obj, err := Generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	_, ok := obj["petId"]
	require.True(t, ok)
}

// TestObj tests generating a random object for body parameters
func TestStruct(t *testing.T) {
	path := "/store/order"
	method := "post"
	obj, err := Generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	dict, ok := obj["body"]
	require.True(t, ok)
	dict2, ok := dict.(map[string]interface{})
	require.True(t, ok)
	_, ok = dict2["complete"]
	require.True(t, ok)
}

func TestArrayWithPrimative(t *testing.T) {
	path := "/pet/findByStatus"
	method := "get"
	obj, err := Generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	_, ok := obj["status"]
	require.True(t, ok)
}

func TestArrayWithObj(t *testing.T) {
	path := "/user/createWithArray"
	method := "post"
	obj, err := Generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	// obj =
	//	{
	//    "body": [
	//        {
	//            "email": "8e981211-9e8e-4267-b985-736241bd578f",
	//            "firstName": "c61c2dc1-e0ad-4a88-881f-f5f9d513fcf4",
	//        }
	//    ]
	//   }
	//util.PrettyPrintStruct(obj)
	val, ok := obj["body"]
	require.True(t, ok)
	//    val = [
	//        {
	//            "email": "8e981211-9e8e-4267-b985-736241bd578f",
	//            "firstName": "c61c2dc1-e0ad-4a88-881f-f5f9d513fcf4",
	//        }
	//          ]
	//util.PrettyPrintStruct(val)
	arr, ok := val.([]interface{})
	require.True(t, ok)
	//util.PrettyPrintStruct(arr)
	items, ok := arr[0].(map[string]interface{})
	require.True(t, ok)
	_, ok = items["email"]
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

func tryOp(op *spec.Operation, method string, path string) {
	if op == nil {
		return
	}
	if len(op.Parameters) == 0 {
		return
	}
	fmt.Printf("Trying %s %s\n", method, path)
	data := map[string]interface{}{}
	for _, param := range op.Parameters {
		data[param.Name] = GenerateAny(&param)
	}
	util.PrettyPrintStruct(op.Parameters)
	util.PrettyPrintStruct(data)
	fmt.Println("**************************************")
}

// TestPetStore mocks the first param for all of pet store
func TestPetStore(t *testing.T) {
	swagger := ReadSwagger(PetStoreExpanded)
	for path, pathItem := range swagger.Paths.Paths {
		tryOp(pathItem.Get, "get", path)
	}
}
