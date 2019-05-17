package swagger

import (
	"fmt"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/util"
)

// ReadSwagger file into memory
func ReadSwagger(path string) *spec.Swagger {
	swagger := &spec.Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}

// findOperation searches a swagger spec for the match path and method, and returns the spec.Operation
func findOperation(swagger *spec.Swagger, key string, method string) (*spec.Operation, error) {
	for path, pathItem := range swagger.Paths.Paths {
		if path == key {
			if method == util.GET {
				return pathItem.Get, nil
			}
			if method == util.DELETE {
				return pathItem.Delete, nil
			}
			if method == util.POST {
				return pathItem.Post, nil
			}
			if method == util.PATCH {
				return pathItem.Patch, nil
			}
			if method == util.PUT {
				return pathItem.Put, nil
			}
			if method == util.HEAD {
				return pathItem.Head, nil
			}
		}
	}
	err := fmt.Errorf("failed to find %v %v in swagger spec", method, key)
	return nil, err
}

// Expand takes a spec, expands it, and writes it to dst
func Expand(spec string, dst string) error {
	doc, err := loads.Spec(spec)
	if err != nil {
		return err
	}
	newDoc, err := doc.Expanded()
	if err != nil {
		return err
	}
	swag := newDoc.Spec()
	return util.MarshalToFile(swag, dst)
}
