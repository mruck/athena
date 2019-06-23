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
