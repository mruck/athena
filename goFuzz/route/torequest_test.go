package route

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/lib/util"
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
	// Populate path params
	route := getRoute(pathTestfile)
	newPath := route.SetPathParams()

	// Build a regexp
	re := regexp.MustCompile(`/\{[^/]+\}`)
	pathRegexp := re.ReplaceAllString(route.Path, "/([^/]+)")
	re, err := regexp.Compile(pathRegexp)

	// Make sure path matches against that
	require.NoError(t, err)
	require.True(t, re.Match([]byte(newPath)))
}

func TestGetQueryStr(t *testing.T) {
	route := getRoute(queryTestfile)
	query := route.GetQueryStr()
	require.True(t, strings.Contains(query, "username"))
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

func handleQuery(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Make sure it has the path
		require.True(t, strings.Contains(req.URL.Path, "/user/login"))
		require.True(t, util.CompareMethods("get", req.Method))
		require.True(t, strings.Contains(req.URL.RawQuery, "username"))
	}
}

func handlePathParam(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Make sure it has the path
		require.True(t, strings.Contains(req.URL.Path, "/pet"))
		// Make sure we replace {}
		require.False(t, strings.Contains(req.URL.Path, "{"))
		require.True(t, util.CompareMethods("get", req.Method))
	}
}

func handleBody(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		require.Equal(t, "/pet", req.URL.Path)
		require.True(t, util.CompareMethods("put", req.Method))
		// Check body
		b, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		dst := map[string]interface{}{}
		err = json.Unmarshal(b, &dst)
		require.NoError(t, err)
		_, ok := dst["category"]
		require.True(t, ok)
		_, ok = dst["id"]
		require.True(t, ok)
	}
}

type handlerCreater func(t *testing.T) http.HandlerFunc

// newServer spins up a new test server
func newServer(t *testing.T, handlerCreater handlerCreater) *httptest.Server {
	return httptest.NewServer(handlerCreater(t))
}

func mockServer(t *testing.T, inputFile string, handlerCreater handlerCreater) {
	// Spin up test server
	ts := newServer(t, handlerCreater)
	defer ts.Close()
	url := urlFromTestServer(t, ts)

	// Allocate httpclient
	client, err := httpclient.New(url)
	require.NoError(t, err)

	// Create our request
	route := getRoute(inputFile)
	req, err := route.ToHTTPRequest()
	require.NoError(t, err)

	// Send it
	_, err = client.Do(req)
	require.NoError(t, err)
}

func TestSendQueryParams(t *testing.T) {
	mockServer(t, queryTestfile, handleQuery)
}

func TestSendBodyParams(t *testing.T) {
	mockServer(t, bodyTestfile, handleBody)
}

func TestSendPathParams(t *testing.T) {
	mockServer(t, pathTestfile, handlePathParam)
}
