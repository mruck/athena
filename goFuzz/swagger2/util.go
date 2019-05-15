package swagger

import (
	"fmt"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/util"
)

// GET for a spec.operation
const GET = "get"

// DELETE for a spec.operation
const DELETE = "delete"

// POST for a spec.operation
const POST = "post"

// PATCH for a spec.operation
const PATCH = "patch"

// PUT for a spec.operation
const PUT = "put"

// HEAD for a spec.operation
const HEAD = "head"

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
			if method == GET {
				return pathItem.Get, nil
			}
			if method == DELETE {
				return pathItem.Delete, nil
			}
			if method == POST {
				return pathItem.Post, nil
			}
			if method == PATCH {
				return pathItem.Patch, nil
			}
			if method == PUT {
				return pathItem.Put, nil
			}
			if method == HEAD {
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
