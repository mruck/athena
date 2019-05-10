package route

import (
	"net/http"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/goFuzz/util"
)

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
