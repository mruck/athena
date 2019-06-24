package swagger

import (
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/mruck/athena/lib/util"
)

// ReadSwagger file into memory
func ReadSwagger(path string) *spec.Swagger {
	swagger := &spec.Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}

// Expand takes a spec, expands it, and writes it to dst
func Expand(spec string, dst string) error {
	doc, err := loads.Spec(spec)
	if err != nil {
		return err
	}
	newDoc, err := doc.Expanded()
	if err != nil {
		return err
	}
	swag := newDoc.Spec()
	return util.MarshalToFile(swag, dst)
}
