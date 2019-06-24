package swagger

// EmbedMetadata embeds metadata object inside top level spec.Parameters,
// with pointers to leaves

import (
	"fmt"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/lib/log"
	"github.com/pkg/errors"
)

const object = "object"
const array = "array"

func embedLeaf(schema *spec.Schema) []*metadata {
	data := newMetadata(*schema)
	//storeSelfReferentialPtr(schema, data)
	return []*metadata{data}
}

func traverseObj(properties *map[string]spec.Schema) []*metadata {
	// We are also storing results to the schema.  Since we can't modify the
	// properties map, allocate a new one
	propertiesPrime := make(map[string]spec.Schema, len(*properties))
	// Keep track of each metadata leaf
	metadataLeaves := []*metadata{}

	for key, schema := range *properties {
		// Explore the children.
		// Hack: pass schema by reference even though its scope is limited to
		// the for loop so that we can modify in place and store shortly after
		// in a newly generated spec.Properties map
		leaves := traverseSchema(&schema)

		// Store the metadata for each child
		metadataLeaves = append(metadataLeaves, leaves...)

		// Create a copy of the new schema since we can't modify the
		// spec.Properties map values
		propertiesPrime[key] = schema
	}

	// Properties should now point to schema with embedded self referential
	// pointers
	*properties = propertiesPrime

	return metadataLeaves
}

func traverseArray(items *spec.SchemaOrArray) []*metadata {
	schema := items.Schema
	if schema == nil {
		err := fmt.Errorf("unhandled: SchemaOrArray is array")
		log.Fatalf("%+v\n", errors.WithStack(err))
	}

	// Array elements are objects
	if schema.Type[0] == object {
		return traverseObj(&schema.Properties)
	}

	// Array elements are primitive, we are in the base case
	return embedLeaf(schema)
}

func traverseSchema(schema *spec.Schema) []*metadata {
	if schema.Type[0] == object {
		return traverseObj(&schema.Properties)
	}
	if schema.Type[0] == array {
		return traverseArray(schema.Items)
	}
	// This is a leaf
	return embedLeaf(schema)
}

// Manipulate a parameter
func embedParam(param *spec.Parameter) {
	// Handle body
	if param.In == "body" {
		// Allocate a metadata object for each leaf, and embed a pointer to it
		metadataLeaves := traverseSchema(param.Schema)
		// Store in a list because its easier to manipulate
		param.VendorExtensible.AddExtension(xmetadata, metadataLeaves)
		return
	}
	// Handle path, header, query, form data.
	// Embed an "x-values" field for storing past values.
	param.VendorExtensible.AddExtension(xmetadata, newMetadata(spec.Schema{}))
}

// Traverse all parameters by operation (i.e. get, put)
func traverseOp(op *spec.Operation) {
	if op == nil {
		return
	}
	for i := range op.Parameters {
		embedParam(&op.Parameters[i])
	}
}

// TraverseSwagger traverse swagger by operation
func TraverseSwagger(swagger *spec.Swagger) {
	for _, pathItem := range swagger.Paths.Paths {
		traverseOp(pathItem.Get)
		traverseOp(pathItem.Delete)
		traverseOp(pathItem.Put)
		traverseOp(pathItem.Patch)
		traverseOp(pathItem.Post)
		traverseOp(pathItem.Head)
	}

}
