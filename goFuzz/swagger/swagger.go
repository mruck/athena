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
func GenerateEnum(enum []interface{}) interface{} {
	randIndex := int(uuid.New().ID()) % len(enum)
	return enum[randIndex]
}

// GenerateBySchema takes a schema and generates random data
// recursively
func GenerateBySchema(schema spec.Schema) interface{} {
	if schema.Enum != nil {
		return GenerateEnum(schema.Enum)
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
	fmt.Printf("data type: %v\n", dataType)
	if dataType == "array" {
		//obj := []interface{}{}
		util.PrettyPrintStruct(*schema.Items)
		//obj[0] = GenerateBySchema(schema.Items)
		//return obj
	}
	return util.Rand(dataType)
}

// GenerateByParam takes a parameter and generates a random value
// according to param.Type
func GenerateByParam(param *spec.Parameter) interface{} {
	// No schema provided
	if param.Schema == nil {
		if param.Type == "object" {
			log.Fatal("Unhandled param type!!!!\n")
		}
		// Do arrays never have schema?
		if param.Type == "array" {
			obj := make([]interface{}, 1)
			if param.Items.Type == "array" {
				log.Fatal("Unhandled param type!!!!\n")
			}
			if param.Items.Type == "object" {
				fmt.Printf("here\n")
				util.PrettyPrintStruct(*param.Items)
				//dict := map[string]interface{}{}
				//for key, schema := range param.Items.Schemas[0].Properties {
				//	dict[key] = GenerateBySchema(schema)
				//}
				//obj[0] = dict
				//return obj
			}
			if param.Items.Enum != nil {
				obj[0] = GenerateEnum(param.Items.Enum)
				return obj
			}
			obj[0] = util.Rand(param.Type)
			return obj
			//util.PrettyPrintStruct(*param.Items)
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
