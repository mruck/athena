package route

import (
	"testing"

	"github.com/mruck/athena/goFuzz/util"
	"github.com/stretchr/testify/require"
)

func TestReadSwagger(t *testing.T) {
	swagger := ReadSwagger("dummySwagger.json")
	//util.PrettyPrintStruct(swagger)
	// Check that a field is correct
	description := swagger.Paths.Paths["/categories.json"].Post.Description
	require.Equal(t, "Create a new category", description)
}

// Just make sure we don't hit an unmarshaling error
func TestReadDiscourseSwagger(t *testing.T) {
	_ = ReadSwagger("discourseSwagger.json")
	//util.PrettyPrintStruct(swagger)
}

type person struct {
	Name string
	Age  int
}

func TestConcatenate(t *testing.T) {
	obj1 := &person{"bob1", 1}
	obj2 := &person{"bob2", 2}
	obj3 := &person{"bob3", 3}
	obj4 := &person{"bob4", 4}
	var list1 = []*person{obj1, obj2}
	var list2 = []*person{obj3, obj4}
	joined := append(list1, list2...)
	util.PrettyPrintStruct(joined)
}
