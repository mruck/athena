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

const xmetadata = "x-metadata"

// custom metadata obj to embed in the leaf node of a swagger parameter
type metadata struct {
	Values []interface{}
	// tainted queries
}

// Allocate a new metadata object
func newMetadata() *metadata {
	return &metadata{Values: []interface{}{}}
}

// Read most recently stored value
func readNewestValue(vendorExtensible *spec.VendorExtensible) interface{} {
	metadata := vendorExtensible.Extensions[xmetadata].(*metadata)
	return metadata.Values[0]
}

// update metadata struct in leaf node of swagger parameter
func updateMetadata(vendorExtensible *spec.VendorExtensible, newVal interface{}) {
	if _, ok := vendorExtensible.Extensions[xmetadata]; !ok {
		// Allocate metadata struct
		vendorExtensible.AddExtension(xmetadata, newMetadata())
	}

	// Cast to a metadata struct
	metadata := vendorExtensible.Extensions[xmetadata].(*metadata)

	// Prepend the new value
	metadata.Values = append([]interface{}{newVal}, metadata.Values...)
}

// GenerateEnum returns a valid enum for the given schema
func GenerateEnum(enum []interface{}) interface{} {
	randIndex := int(uuid.New().ID()) % len(enum)
	return enum[randIndex]
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

// GeneratePrimitiveArray generates an array with only primitive elements
// Query, header, etc params are only allowed arrays with primitives.
func GeneratePrimitiveArray(items *spec.Items) interface{} {
	obj := make([]interface{}, 1)
	if items.Enum != nil {
		obj[0] = GenerateEnum(items.Enum)
		updateMetadata(&items.VendorExtensible, obj)
		return obj
	}
	if items.Type == object {
		err := fmt.Errorf("objects in arrays only allowed in body")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	obj[0] = util.Rand(items.Type)
	return obj
}

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

// GenerateParam runs on all param types except body params
func GenerateParam(param *spec.Parameter) interface{} {
	// We weren't expecting this
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

	// Store the value
	updateMetadata(&param.VendorExtensible, val)
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

// Generate fake parameter data for the first paramater of the given path and method
func Generate(swaggerPath string, path string, method string) (map[string]interface{}, error) {
	swagger := ReadSwagger(swaggerPath)
	op, err := findOperation(swagger, path, method)
	if err != nil {
		return nil, err
	}

	obj := GenerateAny(&op.Parameters[0])

	final := map[string]interface{}{}
	final[op.Parameters[0].Name] = obj
	return final, nil
}

// For testing only.  Load a swagger file and retrieve a parameter
func getParam(swaggerPath string, path string, method string, paramName string) (*spec.Parameter, error) {
	swagger := ReadSwagger(swaggerPath)
	op, err := findOperation(swagger, path, method)
	if err != nil {
		return nil, err
	}
	for i, param := range op.Parameters {
		if param.Name == paramName {
			return &op.Parameters[i], nil
		}
	}
	return nil, nil
}
