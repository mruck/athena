package mutator

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/google/uuid"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/swagger"
	"github.com/mruck/athena/lib/log"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
)

// MutateRoute mutates the parameters on a given route.
// Setting param.Next for each parameter, or nil if the paramater shouldn't
// be sent
func (mutator *Mutator) MutateRoute(route *route.Route) {
	for _, param := range route.Params {
		mutateParam(&param.Parameter)
		// Correctly format the data (i.e. into json)
		param.Next = swagger.Format(&param.Parameter)
	}
}

// mutateEnum returns a valid enum for the given schema
func mutateEnum(enum []interface{}) interface{} {
	randIndex := int(uuid.New().ID()) % len(enum)
	return enum[randIndex]
}

// generate an array with only primitive elements
// Query, header, etc params are only allowed arrays with primitives.
func mutatePrimitiveArray(items *spec.Items) interface{} {
	if items.Type == "object" {
		err := fmt.Errorf("objects in arrays only allowed in body parameters")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}
	// For primitive arrays, these are leaf nodes so we control the number of
	// items in the array
	obj := make([]interface{}, 1)
	if items.Enum != nil {
		obj[0] = mutateEnum(items.Enum)
		return obj
	}
	obj[0] = util.Rand(items.Type)
	return obj
}

// Mutate a primitive parameter (path, query)
func mutatePrimitive(param *spec.Parameter) {
	// Mutate our value
	var val interface{}
	if param.Type == "array" {
		val = mutatePrimitiveArray(param.Items)
	} else if param.Enum != nil {
		val = mutateEnum(param.Enum)
	} else {
		val = util.Rand(param.Type)
	}

	// Store this in the list of values
	swagger.StoreValue(param, val)
}

func mutatePrimitiveSchema(schema spec.Schema) interface{} {
	if schema.Enum != nil {
		return mutateEnum(schema.Enum)
	}
	return util.Rand(schema.Type[0])
}

// Mutate a schema leaf node
func mutateSchema(metadata *swagger.Metadata) {
	schema := metadata.Schema

	// Mutate our value
	var val interface{}
	if schema.Type[0] == "array" {
		obj := make([]interface{}, 1)
		obj[0] = mutatePrimitiveSchema(schema)
		val = obj
	} else {
		val = mutatePrimitiveSchema(schema)
	}

	// Update the metadata object.  This is a pointer so the update
	// is done in place.
	metadata.Values = append([]interface{}{val}, metadata.Values...)
}

// Mutate a body parameter.  At the top level *spec.Parameter, we have a list
// of custom *swagger.Metadata, each representing a leaf in the body.
func mutateBody(param *spec.Parameter) {
	metadatas := swagger.ReadAllMetadata(param)
	for _, metadata := range metadatas {
		// Mutate
		mutateSchema(metadata)
	}
}

func mutateParam(param *spec.Parameter) {
	// This is a multi level object. Mutate the leafs individually.
	if param.In == "body" {
		mutateBody(param)
	} else {
		mutatePrimitive(param)
	}
}
