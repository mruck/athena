package route

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/stretchr/testify/require"
)

const bodyTestfile = "test/body.json"
const pathTestfile = "test/path.json"
const queryTestfile = "test/query.json"

// getRoute returns the first route from the swagger
// file at `path`.  getRoute also mutates the route
// so param.Next is populated
func getRoute(path string) *Route {
	routes := FromSwagger(path)
	routes[0].Mutate()
	return routes[0]
}

func TestSetPathParam(t *testing.T) {
	route := getRoute(pathTestfile)
	path := route.SetPathParams()
	// Make sure it has the path
	require.True(t, strings.Contains(path, "/pet"))
	// Make sure we replace {}
	require.False(t, strings.Contains(path, "{"))
	// TODO: do a check via a regex!
}

func TestGetQueryStr(t *testing.T) {
	route := getRoute(queryTestfile)
	query := route.GetQueryStr()
	// TODO: do a check via a regex!
	require.True(t, strings.Contains(query, "username"))
	fmt.Println(query)
}

func TestGetBodyParams(t *testing.T) {
	route := getRoute(bodyTestfile)
	bodyReader, err := route.GetBodyParams()
	require.NoError(t, err)
	require.NotNil(t, bodyReader)
}

// urlFromTestServer extracts the url from a httptest.Server.
// This is used to create a httpclient.Client
func urlFromTestServer(t *testing.T, ts *httptest.Server) *url.URL {
	url, err := url.Parse(ts.URL)
	require.NoError(t, err)
	return url
}

type correctRequest struct {
	path   string
	method string
}

// newServer spins up a new test server
func newServer(t *testing.T, correct correctRequest) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// TODO: how do i want to check query?
		require.Equal(t, correct.path, req.URL.Path)
		require.Equal(t, correct.method, req.Method)
		// TODO: check body?
	})
	return httptest.NewServer(handler)
}

func TestSendQueryParams(t *testing.T) {
}

func TestSendBodyParams(t *testing.T) {
	correct := correctRequest{path: "/pet", method: "put"}
	// Spin up test server
	ts := newServer(t, correct)
	defer ts.Close()
	url := urlFromTestServer(t, ts)

	// Allocate httpclient
	client, err := httpclient.New(url)
	require.NoError(t, err)

	// Create our request
	route := getRoute(bodyTestfile)
	req, err := route.ToHTTPRequest()
	require.NoError(t, err)

	// Send it
	_, err = client.Do(req)
	require.NoError(t, err)
}

func TestSendPathParams(t *testing.T) {
}
