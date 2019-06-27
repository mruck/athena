package swagger

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
)

// Format the swagger tree structure into a blob of data
func Format(param *spec.Parameter) interface{} {
	// Handle body
	if param.In == "body" {
		return formatSchema(*param.Schema)
	}
	// Handle path, header, query, form data
	return formatParam(param)
}

// formatSchema generates fake data for body parameters
// (i.e. in: body)
func formatSchema(schema spec.Schema) interface{} {
	switch schema.Type[0] {
	case object:
		return formatObj(schema.Properties)
	case array:
		return formatArray(schema.Items)
	default:
		return ReadSchemaValue(schema)
	}
}

// formatObj generates an object
func formatObj(properties map[string]spec.Schema) map[string]interface{} {
	// Allocate object for storing newly formatd data
	obj := map[string]interface{}{}

	for key, schema := range properties {
		obj[key] = formatSchema(schema)
	}

	return obj
}

// formatArray generates an array of any type (including object)
func formatArray(items *spec.SchemaOrArray) []interface{} {
	schema := items.Schema
	if schema == nil {
		err := fmt.Errorf("unhandled: SchemaOrArray is array")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	obj := make([]interface{}, 1)
	if schema.Type[0] == object {
		obj[0] = formatObj(schema.Properties)
		return obj
	}
	obj[0] = ReadSchemaValue(*schema)
	return obj

}

// formatParam runs on all param types except body params
func formatParam(param *spec.Parameter) interface{} {
	// Note: param.Type should never equal "" but sometimes it does cause swagger
	// is human formatd
	if param.Type == object {
		err := fmt.Errorf("unhandled: object in query/header/form data param")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}

	if param.Type == array {
		return ReadParamValue(param)
	}
	return ReadParamValue(param)
}
