package route

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mruck/athena/goFuzz/util"
	"github.com/pkg/errors"
)

// SetPathParams takes the path and populates any path params with param.Next
func (route *Route) SetPathParams() string {
	path := route.Path
	for _, param := range route.State {
		if param.In == "path" {
			// TODO: assert param.Next != nil
			if param.Next == nil {
				err := fmt.Errorf("param %v is nil", param.Name)
				panic(errors.WithStack(err))
			}
			stringified := util.Stringify(param.Next)
			replacer := strings.NewReplacer("{"+param.Name+"}", stringified)
			path = replacer.Replace(path)
		}
	}
	return path
}

// GetQueryStr returns the query string using param.Next
func (route *Route) GetQueryStr() string {
	// Start the query string
	querystr := "?"
	for _, param := range route.State {
		if param.In == "query" {
			// TODO: skip if param.Next == nil
			if param.Next == nil {
				err := fmt.Errorf("param %v is nil", param.Name)
				panic(errors.WithStack(err))
			}
			stringified := util.Stringify(param.Next)
			querystr += param.Name + "=" + stringified + "&"
		}
	}
	// We never added anything
	if querystr == "?" {
		return ""
	}
	// TODO: URL encoding?
	// Strip the trailing &
	return strings.TrimSuffix(querystr, "&")
}

// GetBodyParams marshals body params and returns a reader to those bytes
func (route *Route) GetBodyParams() (io.Reader, error) {
	for _, param := range route.State {
		if param.In == "body" {
			if param.Next == nil {
				err := fmt.Errorf("param %v is nil", param.Name)
				panic(errors.WithStack(err))
			}
			data, err := json.Marshal(param.Next)
			util.Must(err == nil, "%+v\n", errors.WithStack(err))
			// There's only 1 body param, we are done
			return strings.NewReader(string(data)), err
		}
	}
	return nil, nil
}

// ToHTTPRequest converts a route to an http.Request
func (route *Route) ToHTTPRequest() (*http.Request, error) {
	// Populate path parameters
	path := route.SetPathParams()
	url := fmt.Sprintf("http://overwriteMe.com%s", path)

	var body io.Reader

	if util.CompareMethods(route.Method, util.GET) {
		// populate query parameters
		url += route.GetQueryStr()
	} else {
		// set body
		body, _ = route.GetBodyParams()
	}

	req, err := http.NewRequest(strings.ToUpper(route.Method), url, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return req, nil
}
