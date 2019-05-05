package preprocess

// Request contains data for sending a Go request
type Request struct {
	URL     string
	Headers string
}

// HTTPState contains metadata for connecting to a target
type HTTPState struct {
	Host     string
	Port     int
	Requests []Request
}

// GetHTTPState parses a har file with login information and returns
// a series of GO requests to replicate that behavior
func GetHTTPState() *HTTPState {
	return nil
}
