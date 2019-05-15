package swagger

// Documentation:
// https://swagger.io/docs/specification/2-0/describing-parameters/
//
// Nodejs swagger data generator:
// https://github.com/subeeshcbabu/swagmock/blob/master/lib/generators/index.js

import (
	"fmt"
	"log"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/google/uuid"
	"github.com/mruck/athena/goFuzz/util"
	"github.com/pkg/errors"
)

// GenerateEnum returns a valid enum for the given schema
func GenerateEnum(enum []interface{}) interface{} {
	randIndex := int(uuid.New().ID()) % len(enum)
	return enum[randIndex]
}

func GenerateObj(properties map[string]spec.Schema) map[string]interface{} {
	obj := map[string]interface{}{}
	for key, schema := range properties {
		obj[key] = GenerateSchema(schema)
	}
	return obj
}

const object string = "object"

func GenerateArray(items *spec.SchemaOrArray) []interface{} {
	schema := items.Schema
	//util.PrettyPrintStruct(schema)
	if schema == nil {
		err := fmt.Errorf("unhandled: SchemaOrArray is array")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	obj := make([]interface{}, 1)
	if schema.Enum != nil {
		obj[0] = GenerateEnum(schema.Enum)
		return obj
	}
	if schema.Type[0] == object {
		obj[0] = GenerateObj(schema.Properties)
		return obj
	}
	obj[0] = util.Rand(schema.Type[0])
	return obj

}

// GeneratePrimitiveArray generates an array with only primitive elemenets
// Query, header, etc params are only allowed arrays with primitives.
func GeneratePrimitiveArray(items *spec.Items) interface{} {
	obj := make([]interface{}, 1)
	if items.Enum != nil {
		obj[0] = GenerateEnum(items.Enum)
		return obj
	}
	if items.Type == object {
		err := fmt.Errorf("objects in arrays only allowed in body")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	obj[0] = util.Rand(items.Type)
	return obj
}

// GenerateSchema runs on body parameters, i.e in: body
func GenerateSchema(schema spec.Schema) interface{} {
	//util.PrettyPrintStruct(schema)
	//fmt.Println("**************************")
	if schema.Enum != nil {
		return GenerateEnum(schema.Enum)
	}
	if schema.Type[0] == "object" {
		return GenerateObj(schema.Properties)
	}
	if schema.Type[0] == "array" {
		return GenerateArray(schema.Items)
	}
	return util.Rand(schema.Type[0])
}

// GenerateParam runs on all param types except body params
func GenerateParam(param *spec.Parameter) interface{} {
	if param.Enum != nil {
		// TODO: i'm pretty sure this can happen but I want to see a case where it does
		// cause I can't find the enum field of the spec.Parameter struct
		err := fmt.Errorf("unhandled: enum in toplevel non body param")
		log.Fatalf("%+v\n", errors.WithStack(err))
		return GenerateEnum(param.Enum)
	}
	if param.Type == "object" {
		// TODO: Does this make sense for an obj to be in a header/query/etc?
		err := fmt.Errorf("unhandled: object in query/header/form data param")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	if param.Type == "array" {
		return GeneratePrimitiveArray(param.Items)
	}
	return util.Rand(param.Type)
}

// GenerateAny runs on all params, distinguishing between
// body params and all other params
func GenerateAny(param *spec.Parameter) interface{} {
	// Handle body
	if param.In == "body" {
		return GenerateSchema(*param.Schema)
	}
	// Handle path, header, query, form data
	return GenerateParam(param)
}

// ReadSwagger file into memory
func ReadSwagger(path string) *spec.Swagger {
	swagger := &spec.Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}

func Generate(swaggerPath string, path string, method string) (map[string]interface{}, error) {
	swagger := ReadSwagger(swaggerPath)
	op, err := findOperation(swagger, path, method)
	if err != nil {
		return nil, err
	}

	//util.PrettyPrintStruct(op)
	obj := GenerateAny(&op.Parameters[0])
	//util.PrettyPrintStruct(obj)

	final := map[string]interface{}{}
	final[op.Parameters[0].Name] = obj
	//	util.PrettyPrintStruct(final)
	return final, nil
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
