package swagger

import (
	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/util"
)

// Swagger embeds spec.Swagger so I can add custom operations
type Swagger struct {
	spec.Swagger
}

// ReadSwagger file into memory
func ReadSwagger(path string) *Swagger {
	swagger := &Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}

func (swagger *Swagger) findOperation(key string, method string) *spec.Operation {
	for path, pathItem := range swagger.Paths.Paths {
		if path == key {
			if method == "GET" {
				return pathItem.Get
			}
			if method == "Delete" {
				return pathItem.Delete
			}
		}
	}
	return nil
}

// Generate random data for the api with the given path and method
func (swagger *Swagger) Generate(path string, method string) interface{} {
	operation := swagger.findOperation(path, method)
	// recurse on our operation
	return nil
}

// Generatge2 - a slimmer api?
func Generate2(op *spec.Operation) interface{} {
	return nil
}
