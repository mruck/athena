package swagger

// Documentation:
// https://swagger.io/docs/specification/2-0/describing-parameters/
//
// Nodejs swagger data generator:
// https://github.com/subeeshcbabu/swagmock/blob/master/lib/generators/index.js

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/google/uuid"
	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
)

const object = "object"

// GenerateSchema generates fake data for body parameters
// (i.e. in: body)
func GenerateSchema(schema *spec.Schema) interface{} {
	var val interface{}
	switch schema.Type[0] {
	case "object":
		val = GenerateObj(&schema.Properties)
	case "array":
		val = GenerateArray(schema.Items)
	default:
		if schema.Enum != nil {
			val = GenerateEnum(schema.Enum)
		} else {
			val = util.Rand(schema.Type[0])
		}
	}
	return val
}

// GenerateObj generates an object
func GenerateObj(properties *map[string]spec.Schema) map[string]interface{} {
	// Allocate object for storing newly generated data
	obj := map[string]interface{}{}

	// We are also storing results to the schema.  Since we can't modify the properties
	// map, allocate a new one
	propertiesPrime := map[string]spec.Schema{}

	for key, schema := range *properties {
		obj[key] = GenerateSchema(&schema)
		propertiesPrime[key] = schema
	}

	*properties = propertiesPrime

	return obj
}

// GenerateArray generates an array of any type (including object)
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
		obj[0] = GenerateObj(&schema.Properties)
		return obj
	}
	obj[0] = util.Rand(schema.Type[0])
	return obj

}

// GenerateEnum returns a valid enum for the given schema
func GenerateEnum(enum []interface{}) interface{} {
	randIndex := int(uuid.New().ID()) % len(enum)
	return enum[randIndex]
}

// GeneratePrimitiveArray generates an array with only primitive elements
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

// GenerateParam runs on all param types except body params
func GenerateParam(param *spec.Parameter) interface{} {
	// Note: param.Type should never equal "" but sometimes it does cause swagger
	// is human generated
	if param.Type == "object" {
		err := fmt.Errorf("unhandled: object in query/header/form data param")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}

	// Generate a value
	var val interface{}
	if param.Type == "array" {
		val = GeneratePrimitiveArray(param.Items)
	} else {
		if param.Enum != nil {
			val = GenerateEnum(param.Enum)
		} else {
			val = util.Rand(param.Type)
		}
	}

	return val
}

// GenerateAny generates fake data for all param types
func GenerateAny(param *spec.Parameter) interface{} {
	// Handle body
	if param.In == "body" {
		return GenerateSchema(param.Schema)
	}
	// Handle path, header, query, form data
	return GenerateParam(param)
}
