package swagger

import "github.com/go-openapi/spec"

const xmetadata = "x-metadata"
const xreferential = "x-self-referential"

// Metadata obj to embed at the top level of a parameter.  This is used
// to set next values and store past values.  Multi level parameters
// store pointers to this at the leaf level and read the next value from here
type metadata struct {
	// Store past and present values
	Values []interface{}
	// Store a copy of the leaf for multi level data structures
	Schema spec.Schema
	// tainted queries
}

// Allocate a new metadata object
func newMetadata(schema spec.Schema) *metadata {
	return &metadata{Values: []interface{}{},
		Schema: schema}
}

func storeSelfReferentialPtr(schema *spec.Schema, ptr *metadata) {
	schema.VendorExtensible.AddExtension(xreferential, ptr)
}

//func storeSelfReferentialPtr(vendorExtensible *spec.VendorExtensible, ptr *metadata) {
//	vendorExtensible.AddExtension(xmetadata, newMetadata(nil))
//	if _, ok := vendorExtensible.Extensions[xmetadata]; !ok {
//		// Allocate metadata struct
//		vendorExtensible.AddExtension(xmetadata, newMetadata(nil))
//	}
//
//	// Cast to a metadata struct
//	metadata := vendorExtensible.Extensions[xmetadata].(*metadata)
//
//	// Prepend the new value
//	metadata.Values = append([]interface{}{newVal}, metadata.Values...)
//}
//// Read most recently stored value
//func readNewestValue(vendorExtensible *spec.VendorExtensible) interface{} {
//	metadata := vendorExtensible.Extensions[xmetadata].(*metadata)
//	return metadata.Values[0]
//}
