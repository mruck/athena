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
	obj, err := mock(PetStoreExpanded, path, method)
	require.NoError(t, err)
	_, ok := obj["petId"]
	require.True(t, ok)
}

// TestObj tests generating a random object for body parameters
func TestStruct(t *testing.T) {
	path := "/store/order"
	method := "post"
	obj, err := mock(PetStoreExpanded, path, method)
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
	obj, err := mock(PetStoreExpanded, path, method)
	require.NoError(t, err)
	_, ok := obj["status"]
	require.True(t, ok)
}

func TestArrayWithObj(t *testing.T) {
	path := "/user/createWithArray"
	method := "post"
	obj, err := mock(PetStoreExpanded, path, method)
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
		data[param.Name] = MockAny(&param)
	}
}

// TestPetStore mocks params for all of pet store
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

// TestDiscourse mocks params for all of discourse
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
