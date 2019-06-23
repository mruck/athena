package swagger

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
)

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
	updateMetadata(&schema.VendorExtensible, val)
	//util.PrettyPrintStruct(schema, nil)
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
		updateMetadata(&schema.VendorExtensible, obj)
		return obj
	}
	if schema.Type[0] == object {
		obj[0] = GenerateObj(&schema.Properties)
		return obj
	}
	obj[0] = util.Rand(schema.Type[0])
	return obj

}
