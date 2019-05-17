package preprocess

import (
	"testing"

	"github.com/mruck/athena/goFuzz/har"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/stretchr/testify/require"
)

func TestHarToRequest(t *testing.T) {
	har := har.UnmarshalHar("../tests/login_har.json")
	requests, err := har.ToHTTPRequests()
	require.NoError(t, err)
	require.NotNil(t, requests)
	// TODO: check headers?
	// TODO: check body?
}
func TestCorpus(t *testing.T) {
	harData := har.UnmarshalHar("../tests/corpus_har.json")
	routes := route.FromSwagger("../swagger/test/discourseSwagger.json")
	_, err := InitializeRoutes(routes, harData)
	require.NoError(t, err)
	//	for _, route := range corpus {
	//		fmt.Printf("%v %v\n", route.Method, route.Path)
	//	}

}
