package swagger

import (
	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/util"
)

func ReadSwagger(path string) *spec.Swagger {
	swagger := &spec.Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}
