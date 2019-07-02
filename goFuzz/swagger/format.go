package swagger

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
)

// Format the swagger tree structure into a blob of data
//func Format(param *spec.Parameter) map[string]interface{} {
func Format(param *spec.Parameter) interface{} {
	if param.In == "body" {
		return formatSchema(*param.Schema)
	}
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
	// Arrays with objects as elements are never leaf nodes, so
	// format controls the number of items in the elements.  Default
	// to 1 element. Eventually we should do a second pass for array
	// sizing.
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
	return ReadParamValue(param)
}
