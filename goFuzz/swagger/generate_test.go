package swagger

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

const PetStoreExpanded = "test/petstore_expanded.json"
const PetStore = "test/petstore.json"

// TestPathParam tests a path parameter
func TestPathParam(t *testing.T) {
	path := "/pet/{petId}"
	method := "get"
	obj, err := generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	_, ok := obj["petId"]
	require.True(t, ok)
}

// TestObj tests generating a random object for body parameters
func TestStruct(t *testing.T) {
	path := "/store/order"
	method := "post"
	obj, err := generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	dict, ok := obj["body"]
	require.True(t, ok)
	dict2, ok := dict.(map[string]interface{})
	require.True(t, ok)
	_, ok = dict2["complete"]
	require.True(t, ok)
}

//func TestMetaPrimitive(t *testing.T) {
//	// Get the param obj
//	path := "/pet/{petId}"
//	method := "get"
//	paramName := "petId"
//	param, err := getParam(PetStoreExpanded, path, method, paramName)
//	require.NoError(t, err)
//
//	// Generate an arbitrary value
//	value := generateParam(param)
//
//	// Make sure it matches the value we stored
//	stored := readNewestValue(&param.VendorExtensible)
//	require.Equal(t, value, stored)
//}
//
//// Test storing a struct as metadata
//func TestMetaStruct(t *testing.T) {
//
//	// Generate param data
//	path := "/store/order"
//	method := "post"
//	paramName := "body"
//	param, err := getParam(PetStoreExpanded, path, method, paramName)
//	require.NoError(t, err)
//	val := GenerateAny(param)
//
//	// Check the entire body stored vs generated
//	stored := readNewestValue(&param.Schema.VendorExtensible)
//	require.Equal(t, val, stored)
//
//	// Check one of the stored leaf nodes
//	schema := param.Schema.Properties["complete"]
//	val = readNewestValue(&schema.VendorExtensible)
//	require.NotNil(t, val)
//}
//
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

func TestArrayWithPrimative(t *testing.T) {
	path := "/pet/findByStatus"
	method := "get"
	obj, err := generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	_, ok := obj["status"]
	require.True(t, ok)
}

func TestArrayWithObj(t *testing.T) {
	path := "/user/createWithArray"
	method := "post"
	obj, err := generate(PetStoreExpanded, path, method)
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
	val, ok := obj["body"]
	require.True(t, ok)
	//    val = [
	//        {
	//            "email": "8e981211-9e8e-4267-b985-736241bd578f",
	//            "firstName": "c61c2dc1-e0ad-4a88-881f-f5f9d513fcf4",
	//        }
	//          ]
	arr, ok := val.([]interface{})
	require.True(t, ok)
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

var broken = map[string]string{
	"post /admin/api/web_hooks": "",
}

func tryOp(op *spec.Operation, method string, path string) {
	if op == nil {
		return
	}
	if len(op.Parameters) == 0 {
		return
	}
	key := method + " " + path
	if _, ok := broken[key]; ok {
		return
	}
	data := map[string]interface{}{}
	//log.Infof("**************************************")
	//log.Infof("Trying %s %s\n", method, path)
	for _, param := range op.Parameters {
		//log.Info(param.Name)
		data[param.Name] = GenerateAny(&param)
	}
}

// TestPetStore generates params for all of pet store
func TestPetStore(t *testing.T) {
	swagger := ReadSwagger(PetStoreExpanded)
	for path, pathItem := range swagger.Paths.Paths {
		tryOp(pathItem.Get, "get", path)
		tryOp(pathItem.Delete, "delete", path)
		tryOp(pathItem.Put, "put", path)
		tryOp(pathItem.Patch, "patch", path)
		tryOp(pathItem.Post, "post", path)
		tryOp(pathItem.Head, "head", path)
	}
}

const discourseSwagger = "test/discourseSwagger.json"

// TestDiscourse generates params for all of discourse
func TestDiscourse(t *testing.T) {
	swagger := ReadSwagger(discourseSwagger)
	for path, pathItem := range swagger.Paths.Paths {
		tryOp(pathItem.Get, "get", path)
		tryOp(pathItem.Delete, "delete", path)
		tryOp(pathItem.Put, "put", path)
		tryOp(pathItem.Patch, "patch", path)
		tryOp(pathItem.Post, "post", path)
		tryOp(pathItem.Head, "head", path)
	}
}

func TestReadSwagger(t *testing.T) {
	swagger := ReadSwagger("../tests/dummySwagger.json")
	// Check that a field is correct
	description := swagger.Paths.Paths["/categories.json"].Post.Description
	require.Equal(t, "Create a new category", description)
}

// Just make sure we don't hit an unmarshaling error
func TestReadDiscourseSwagger(t *testing.T) {
	_ = ReadSwagger("../tests/discourseSwagger.json")
}
