package route

import "net/http"

// Route object contains metadata about a route
type Route struct {
	Path            string
	Method          string
	QueryParams     string
	DynamicSegments string
	BodyParams      string
}

// LoadRoutes reads routes from shared mount and loads them into memory
func LoadRoutes() []*Route {
	return nil
}

// ToHTTPRequest converts a route to an http.Request
func (route *Route) ToHTTPRequest() *http.Request {
	return nil
}
