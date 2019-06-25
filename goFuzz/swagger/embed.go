package swagger

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
)

const object = "object"
const array = "array"

func embedLeaf(schema *spec.Schema) []*Metadata {
	data := newMetadata(*schema)
	embedSelfReferentialPtr(schema, data)
	return []*Metadata{data}
}

func embedObj(properties *map[string]spec.Schema) []*Metadata {
	// We are also storing results to the schema.  Since we can't modify the
	// properties map, allocate a new one
	propertiesPrime := make(map[string]spec.Schema, len(*properties))
	// Keep track of each Metadata leaf
	MetadataLeaves := []*Metadata{}

	for key, schema := range *properties {
		// Explore the children.
		// Hack: pass schema by reference even though its scope is limited to
		// the for loop so that we can modify in place and store shortly after
		// in a newly generated spec.Properties map
		leaves := embedSchema(&schema)

		// Store the Metadata for each child
		MetadataLeaves = append(MetadataLeaves, leaves...)

		// Create a copy of the new schema since we can't modify the
		// spec.Properties map values
		propertiesPrime[key] = schema
	}

	// Properties should now point to schema with embedded self referential
	// pointers
	*properties = propertiesPrime

	return MetadataLeaves
}

func embedArray(items *spec.SchemaOrArray) []*Metadata {
	schema := items.Schema
	if schema == nil {
		err := fmt.Errorf("unhandled: SchemaOrArray is array")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}

	// Array elements are objects
	if schema.Type[0] == object {
		return embedObj(&schema.Properties)
	}

	// Array elements are primitive, we are in the base case
	return embedLeaf(schema)
}

func embedSchema(schema *spec.Schema) []*Metadata {
	if schema.Type[0] == object {
		return embedObj(&schema.Properties)
	}
	if schema.Type[0] == array {
		return embedArray(schema.Items)
	}
	// This is a leaf
	return embedLeaf(schema)
}

// EmbedParam embeds a list of Metadata objects inside a
// top level parameter.  Each Metadata object contains
// past values of a leaf node, and a copy of a leaf node.
// If this is a path/query param, this is a singleton
// list and there's no copy of the leaf node because
// the top level parameter is the leaf node.
func EmbedParam(param *spec.Parameter) {
	// Handle body
	if param.In == "body" {
		// Allocate a Metadata object for each leaf, and embed a pointer to it
		MetadataLeaves := embedSchema(param.Schema)
		// Store in a list because its easier to manipulate
		embedMetadata(param, MetadataLeaves)
		return
	}
	// Handle path, header, query, form data.
	embedMetadata(param, nil)
}
