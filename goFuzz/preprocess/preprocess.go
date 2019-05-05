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

// Corpus contains Go formated requests to use as initial corpus
type Corpus struct {
	Requests []Request
}

// GetCorpus parses Har file, formating relevant info like url, headers, params,
// etc and formating into a list of requests
func GetCorpus() *Corpus {
	return nil
}
