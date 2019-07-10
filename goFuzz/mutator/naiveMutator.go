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
		mutator.mutateParam(&param.Parameter)

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
	maxItems := 100000
	obj := make([]interface{}, maxItems)
	if items.Enum != nil {
		for i := 0; i < maxItems; i++ {
			obj[i] = mutateEnum(items.Enum)
		}
	} else {
		for i := 0; i < maxItems; i++ {
			obj[i] = util.Rand(items.Type)
		}
	}

	return obj
}

func mutatePrimitiveSchema(schema spec.Schema) interface{} {
	if schema.Enum != nil {
		return mutateEnum(schema.Enum)
	}
	return util.Rand(schema.Type[0])
}

// Mutate a schema leaf node
func mutateSchema(metadata *swagger.Metadata) interface{} {
	schema := metadata.Schema

	// Mutate our value
	var val interface{}
	if schema.Type[0] == "array" {
		maxItems := 100000
		obj := make([]interface{}, maxItems)
		for i := 0; i < maxItems; i++ {
			obj[i] = mutatePrimitiveSchema(schema)
		}
		val = obj
	} else {
		val = mutatePrimitiveSchema(schema)
	}
	return val
}

func (mutator *Mutator) mutateTaintedQuery(metadata *swagger.Metadata) interface{} {
	// No queries associated with this param
	if metadata.TaintedQuery == nil {
		return nil
	}

	// Look up a value
	val := mutator.DB.Conn.LookUp(metadata.TaintedQuery.Table, metadata.TaintedQuery.Column)

	// We've never sent this value
	if !util.Contains(metadata.Values, val) {
		return val
	}

	// Stringify and concatenate a semicolon
	stringified := ";" + util.Stringify(val)
	if !util.Contains(metadata.Values, stringified) {
		return stringified
	}

	return val
}

// Mutate a body parameter.  At the top level *spec.Parameter, we have a list
// of custom *swagger.Metadata, each representing a leaf in the body.
func (mutator *Mutator) mutateBody(param *spec.Parameter) {
	metadatas := swagger.ReadAllMetadata(param)
	for _, metadata := range metadatas {
		// Try query based mutation
		val := mutator.mutateTaintedQuery(metadata)

		// Query based mutation failed
		if val == nil {
			// Mutate
			val = mutateSchema(metadata)
		}

		// Update the metadata object.  This is a pointer so the update
		// is done in place.
		metadata.Values = append([]interface{}{val}, metadata.Values...)
	}
}

// Mutate a primitive parameter (path, query)
func (mutator *Mutator) mutatePrimitive(param *spec.Parameter) {
	var val interface{}

	// Try query based mutation
	metadata := swagger.ReadOneMetadata(param)
	val = mutator.mutateTaintedQuery(metadata)

	// We failed to use query based mutation
	if val == nil {
		if param.Type == "array" {
			val = mutatePrimitiveArray(param.Items)
		} else if param.Enum != nil {
			val = mutateEnum(param.Enum)
		} else {
			val = util.Rand(param.Type)
		}
	}

	// Update the metadata object
	metadata.Values = append([]interface{}{val}, metadata.Values...)
}

func (mutator *Mutator) mutateParam(param *spec.Parameter) {
	// This is a multi level object. Mutate the leafs individually.
	if param.In == "body" {
		mutator.mutateBody(param)
	} else {
		mutator.mutatePrimitive(param)
	}
}
