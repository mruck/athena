package route

import (
	"net/http"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/param"
	"github.com/mruck/athena/goFuzz/util"
)

// Route contains static information from a swagger, and dynamic mutation state
type Route struct {
	// key in Spec.Swagger.Paths.Paths
	Path string
	// field in Spec.Swagger.Path.Paths[Path].Get/Put/etc
	Method string
	// value for field in Spec.Swagger.Path.Path[Path].Get
	Meta  *spec.Operation
	State *param.State
}

// ReadSwagger file into memory
func ReadSwagger(path string) *spec.Swagger {
	swagger := &spec.Swagger{}
	util.MustUnmarshalFile(path, swagger)
	return swagger
}

// ToHTTPRequest converts a route to an http.Request
func (route *Route) ToHTTPRequest() (*http.Request, error) {
	return nil, nil
	//url := fmt.Sprintf("http://overwriteMe.com%s", route.Path)
	//req, err := http.NewRequest(route.Method, url, nil)
	//if err != nil {
	//	return nil, errors.Wrap(err, "")
	//}
	//return req, nil
}
