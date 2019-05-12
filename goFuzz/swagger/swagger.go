package swagger

import (
	"fmt"
	"log"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/google/uuid"
	"github.com/mruck/athena/goFuzz/util"
)

// ReadSwagger file into memory
func ReadSwagger(path string) *spec.Swagger {
	swagger := &spec.Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}

func findOperation(swagger *spec.Swagger, key string, method string) (*spec.Operation, error) {
	for path, pathItem := range swagger.Paths.Paths {
		if path == key {
			if method == "get" {
				return pathItem.Get, nil
			}
			if method == "delete" {
				return pathItem.Delete, nil
			}
			if method == "post" {
				return pathItem.Post, nil
			}
		}
	}
	err := fmt.Errorf("failed to find %v %v in swagger spec", method, key)
	return nil, err
}

// GenerateEnum returns a valid enum for the given schema
func GenerateEnum(schema spec.Schema) interface{} {
	randIndex := int(uuid.New().ID()) % len(schema.Enum)
	return schema.Enum[randIndex]
}

// GenerateBySchema takes a schema and generates random data
// recursively
func GenerateBySchema(schema spec.Schema) interface{} {
	if schema.Enum != nil {
		return GenerateEnum(schema)
	}

	//util.PrettyPrintStruct(schema)
	dataType := schema.Type[0]
	if dataType == "object" {
		obj := map[string]interface{}{}
		for key, schema := range schema.Properties {
			obj[key] = GenerateBySchema(schema)
		}
		return obj
	}
	//if dataType == "array" {
	//	obj := []interface{}{}
	//	obj[0] = mutateBySchema(schema.Items.Properties)
	//	return obj
	//}
	return util.Rand(dataType)
}

// GenerateByParam takes a parameter and generates a random value
// according to param.Type
func GenerateByParam(param *spec.Parameter) interface{} {
	// No schema provided
	if param.Schema == nil {
		if param.Type == "object" || param.Type == "array" {
			log.Fatal("Unhandled param type\n")
		}
		return util.Rand(param.Type)
	}
	return GenerateBySchema(*param.Schema)
}

// Generate random data for the api with the given path and method
func Generate(spec string, path string, method string) (map[string]interface{}, error) {
	swagger := ReadSwagger(spec)
	op, err := findOperation(swagger, path, method)
	if err != nil {
		return nil, err
	}
	//util.PrettyPrintStruct(*op)
	obj := GenerateByParam(&op.Parameters[0])
	final := map[string]interface{}{}
	final[op.Parameters[0].Name] = obj
	return final, nil
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
