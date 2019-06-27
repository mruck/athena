package swagger

// The Metadata struct is the solution to giving granular control over complex
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

import (
	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/sql"
)

const xmetadata = "x-metadata"
const xreferential = "x-self-referential"

// Metadata obj to embed at the top level of a parameter.  This is used
// to set next values and store past values.  Multi level parameters
// store pointers to this at the leaf level and read the next value from here
type Metadata struct {
	// Store past and present values
	Values []interface{}
	// Store a copy of the leaf for multi level data structures.
	// Ignore this for primitive params i.e. path, query
	Schema spec.Schema
	// For now, only support one query per param, but eventually we should
	// either intelligently merge queries or support multiple queries
	TaintedQuery sql.TaintedQuery
}

// ReadSchemaValue extract the metadata ptr embedded in the schema and reads
// the most recently stored value
// embedded in a schema
func ReadSchemaValue(schema spec.Schema) interface{} {
	// Extract the embedded metadata struct
	metadata := schema.VendorExtensible.Extensions[xreferential].(*Metadata)
	return metadata.Values[0]
}

// ReadParamValue reads the most recently stored value in the first metadata object
// embedded in a param.  This is to be called on path/query params only.
// Body params will return a list of values for each leaf node.
func ReadParamValue(param *spec.Parameter) interface{} {
	// Extract the embedded metadata struct
	metadata := param.VendorExtensible.Extensions[xmetadata].([]*Metadata)[0]
	return metadata.Values[0]
}

// StoreValue stores a single value in the metadata object embedded.  Stores
// should only ever be done from top level spec.Paramter.  The pointer
// to the embedded metadata struct in the leaf nodes is read only.
func StoreValue(param *spec.Parameter, val interface{}) {
	// Extract the embedded metadata struct
	metadata := ReadOneMetadata(param)
	// Insert the new value into index 0.  This is a pointer so the update
	// is done in place.
	metadata.Values = append([]interface{}{val}, metadata.Values...)
}

// ReadOneMetadata reads a single Metadata object.
// This should be called for query/path params where we only have one Metadata obj
func ReadOneMetadata(param *spec.Parameter) *Metadata {
	return param.VendorExtensible.Extensions[xmetadata].([]*Metadata)[0]
}

// ReadAllMetadata reads the metadata extension in the top level parameter.
// This contains metadata objects for every leaf node.
// For non-body params, this is a singleton list
func ReadAllMetadata(param *spec.Parameter) []*Metadata {
	return param.VendorExtensible.Extensions[xmetadata].([]*Metadata)
}

// UpdateMetadata inside spec.Parameter.  The metadata object is stored in a map
// so the only way to update it is by overwriting the old one.
func UpdateMetadata(param *spec.Parameter, metadata *Metadata) {
	param.VendorExtensible.AddExtension(xmetadata, metadata)
}

// Allocate a new Metadata object
func newMetadata(schema spec.Schema) *Metadata {
	return &Metadata{
		Values: []interface{}{},
		Schema: schema,
	}
}

// Embed Metadata in top level parameter.  If MetadataLeaves is nil,
// this is a query/path param so initialize an empty Metadata struct
// for storing stuff later
func embedMetadata(param *spec.Parameter, MetadataLeaves []*Metadata) {
	if MetadataLeaves == nil {
		// Allocate an empty meta data obj
		meta := newMetadata(spec.Schema{})
		MetadataLeaves = []*Metadata{meta}
	}
	param.VendorExtensible.AddExtension(xmetadata, MetadataLeaves)
}

// Embed a pointer to Metadata obj in the leaf node.  The Metadata obj is
// mutated and read from during data generation.  The tree structure
// is only preserved for structuring the data correctly.
func embedSelfReferentialPtr(schema *spec.Schema, ptr *Metadata) {
	schema.VendorExtensible.AddExtension(xreferential, ptr)
}

// Read the most recent value from the leaf node.  This is to be called
// when traversing the tree from the leaf node to get the next value.
func readNextValue(schema *spec.Schema) interface{} {
	meta := schema.VendorExtensible.Extensions[xreferential].(*Metadata)
	return meta.Values[0]
}
