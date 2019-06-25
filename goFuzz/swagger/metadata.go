package swagger

// The metadata struct is the solution to giving granular control over complex
// body params.  Path/query params can be trivially mutated because they are
// only a single level deep.  But body params can be complex json blobs, and
// we want to have the ability to mutate each leaf node individually.  Originally,
// I wanted to store a list of pointers to leaf nodes and mutate those.
// However, we can't have pointers to values in maps and the spec.Schema objects
// are usually values in the spec.Properties map.  So instead, at the top level
// spec.Parameter we keep a list that contains a copy of each leaf node and we
// mutate those.  We embed a pointer inside each leaf node to point to this copy.
// That way, on muation, the leaf node reads the next values from this mutated
// copy of itself.

import "github.com/go-openapi/spec"

const xmetadata = "x-metadata"
const xreferential = "x-self-referential"

// TODO: Maybe all these functions should be methods on our custom param object?

// Metadata obj to embed at the top level of a parameter.  This is used
// to set next values and store past values.  Multi level parameters
// store pointers to this at the leaf level and read the next value from here
type metadata struct {
	// Store past and present values
	Values []interface{}
	// Store a copy of the leaf for multi level data structures.
	// Ignore this for primitive params i.e. path, query
	Schema spec.Schema
	// tainted queries
}

// Allocate a new metadata object
func newMetadata(schema spec.Schema) *metadata {
	return &metadata{Values: []interface{}{},
		Schema: schema}
}

// Embed metadata in top level parameter.  If metadataLeaves is nil,
// this is a query/path param so initialize an empty metadata struct
// for storing stuff later
func embedMetadata(param *spec.Parameter, metadataLeaves []*metadata) {
	if metadataLeaves == nil {
		// Allocate an empty meta data obj
		meta := newMetadata(spec.Schema{})
		metadataLeaves = []*metadata{meta}
	}
	param.VendorExtensible.AddExtension(xmetadata, metadataLeaves)
}

// Embed a pointer to metadata obj in the leaf node.  The metadata obj is
// mutated and read from during data generation.  The tree structure
// is only preserved for structuring the data correctly.
func storeSelfReferentialPtr(schema *spec.Schema, ptr *metadata) {
	schema.VendorExtensible.AddExtension(xreferential, ptr)
}

// Read all past values
func readValues(param *spec.Parameter) []interface{} {
	metadata := param.VendorExtensible.Extensions[xmetadata].(*metadata)
	return metadata.Values
}

// Read metadata extension in the top level parameter.  This contains
// metadata objects for every leaf node. For non-body params, this is a
// singleton list
func readMetadata(param *spec.Parameter) []*metadata {
	return param.VendorExtensible.Extensions[xmetadata].([]*metadata)
}

// Read a single metadata object.  This should be called for query/path params
// where we only have one metadata obj
func readOneMetadata(param *spec.Parameter) *metadata {
	return param.VendorExtensible.Extensions[xmetadata].([]*metadata)[0]
}

// Read the most recent value from the leaf node.  This is to be called
// when traversing the tree from the leaf node to get the next value.
func readNextValue(schema *spec.Schema) interface{} {
	meta := schema.VendorExtensible.Extensions[xreferential].(*metadata)
	return meta.Values[0]
}
