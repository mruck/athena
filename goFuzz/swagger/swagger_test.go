package swagger

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const PetStore = "tests/petstore.json"
const PetStoreExpanded = "tests/petstore_expanded.json"

//const PetStoreSchema = "tests/petstore_schema.json"

//func TestSchema(t *testing.T) {
//	data := &spec.Schema{}
//	util.MustUnmarshalFile(PetStoreSchema, data)
//	//util.PrettyPrintStruct(data)
//	randObj := GenerateBySchema(*data)
//	util.PrettyPrintStruct(randObj)
//}

func TestGenerate(t *testing.T) {
	path := "/pet/{petId}"
	method := "get"
	obj, err := Generate(PetStoreExpanded, path, method)
	require.NoError(t, err)
	fmt.Printf("%v\n", obj)
}

func TestExpandSchema(t *testing.T) {
	err := Expand(PetStore, PetStoreExpanded)
	require.NoError(t, err)
}
