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

// generateSchema generates fake data for body parameters
// (i.e. in: body)
func generateSchema(schema *spec.Schema) interface{} {
	var val interface{}
	switch schema.Type[0] {
	case "object":
		val = generateObj(&schema.Properties)
	case "array":
		val = generateArray(schema.Items)
	default:
		if schema.Enum != nil {
			val = generateEnum(schema.Enum)
		} else {
			val = util.Rand(schema.Type[0])
		}
	}
	return val
}

// generateObj generates an object
func generateObj(properties *map[string]spec.Schema) map[string]interface{} {
	// Allocate object for storing newly generated data
	obj := map[string]interface{}{}

	// We are also storing results to the schema.  Since we can't modify the properties
	// map, allocate a new one
	propertiesPrime := map[string]spec.Schema{}

	for key, schema := range *properties {
		obj[key] = generateSchema(&schema)
		propertiesPrime[key] = schema
	}

	*properties = propertiesPrime

	return obj
}

// generateArray generates an array of any type (including object)
func generateArray(items *spec.SchemaOrArray) []interface{} {
	schema := items.Schema
	if schema == nil {
		err := fmt.Errorf("unhandled: SchemaOrArray is array")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	obj := make([]interface{}, 1)
	if schema.Enum != nil {
		obj[0] = generateEnum(schema.Enum)
		return obj
	}
	if schema.Type[0] == object {
		obj[0] = generateObj(&schema.Properties)
		return obj
	}
	obj[0] = util.Rand(schema.Type[0])
	return obj

}

// generateEnum returns a valid enum for the given schema
func generateEnum(enum []interface{}) interface{} {
	randIndex := int(uuid.New().ID()) % len(enum)
	return enum[randIndex]
}

// generatePrimitiveArray generates an array with only primitive elements
// Query, header, etc params are only allowed arrays with primitives.
func generatePrimitiveArray(items *spec.Items) interface{} {
	obj := make([]interface{}, 1)
	if items.Enum != nil {
		obj[0] = generateEnum(items.Enum)
		return obj
	}
	if items.Type == object {
		err := fmt.Errorf("objects in arrays only allowed in body")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	obj[0] = util.Rand(items.Type)
	return obj
}

// generateParam runs on all param types except body params
func generateParam(param *spec.Parameter) interface{} {
	// Note: param.Type should never equal "" but sometimes it does cause swagger
	// is human generated
	if param.Type == "object" {
		err := fmt.Errorf("unhandled: object in query/header/form data param")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}

	if param.Type == "array" {
		return generatePrimitiveArray(param.Items)
	}
	if param.Enum != nil {
		return generateEnum(param.Enum)
	}
	return util.Rand(param.Type)
}

// GenerateAny generates fake data for all param types
func GenerateAny(param *spec.Parameter) interface{} {
	// Handle body
	if param.In == "body" {
		return generateSchema(param.Schema)
	}
	// Handle path, header, query, form data
	return generateParam(param)
}
