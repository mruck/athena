package swagger

import "github.com/go-openapi/spec"

const xmetadata = "x-metadata"

// custom metadata obj to embed in the leaf node of a swagger parameter
type metadata struct {
	Values []interface{}
	// tainted queries
}

// Allocate a new metadata object
func newMetadata() *metadata {
	return &metadata{Values: []interface{}{}}
}

// Read most recently stored value
func readNewestValue(vendorExtensible *spec.VendorExtensible) interface{} {
	metadata := vendorExtensible.Extensions[xmetadata].(*metadata)
	return metadata.Values[0]
}

// update metadata struct in leaf node of swagger parameter
func updateMetadata(vendorExtensible *spec.VendorExtensible, newVal interface{}) {
	if _, ok := vendorExtensible.Extensions[xmetadata]; !ok {
		// Allocate metadata struct
		vendorExtensible.AddExtension(xmetadata, newMetadata())
	}

	// Cast to a metadata struct
	metadata := vendorExtensible.Extensions[xmetadata].(*metadata)

	// Prepend the new value
	metadata.Values = append([]interface{}{newVal}, metadata.Values...)
}
