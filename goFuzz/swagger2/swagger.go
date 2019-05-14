package swagger

import (
	"fmt"
	"log"

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
	}
	obj[0] = util.Rand(schema.Type[0])
	return obj

}

func GeneratePrimativeArray(items *spec.Items) interface{} {
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

func GenerateSchema(schema spec.Schema) interface{} {
	if schema.Enum != nil {
		// TODO: does it make sense for enum to be top level?
		err := fmt.Errorf("unhandled: enum in toplevel schema")
		log.Fatalf("%+v\n", errors.WithStack(err))
		//return GenerateEnum(schema.Enum)
	}
	if schema.Type[0] == "object" {
		return GenerateObj(schema.Properties)
	}
	if schema.Type[0] == "array" {
		return GenerateArray(schema.Items)
	}
	err := fmt.Errorf("unhandled: schema with primative type")
	log.Fatalf("%+v\n", errors.WithStack(err))
	return util.Rand(schema.Type[0])
}

// GenerateVal runs on everything but a schema
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
		err := fmt.Errorf("unhandled obj")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	if param.Type == "array" {
		return GeneratePrimativeArray(param.Items)
	}
	return util.Rand(param.Type)
}

func Generate(param *spec.Parameter) interface{} {
	// Handle body
	if param.In == "body" {
		_ = GenerateSchema(*param.Schema)
	}
	// Handle path, header, query, form data
	return GenerateParam(param)
}
