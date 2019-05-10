package preprocess

import (
	"fmt"
	"testing"

	"github.com/mruck/athena/goFuzz/route"
	"github.com/stretchr/testify/require"
)

func TestHarToRequest(t *testing.T) {
	har := unmarshalHar("test/login_har.json")
	requests, err := har.toRequests()
	require.NoError(t, err)
	require.NotNil(t, requests)
	// TODO: check headers?
	// TODO: check body?
}
func TestCorpus(t *testing.T) {
	har := unmarshalHar("test/corpus_har.json")
	routes := route.FromSwagger("../route/discourseSwagger.json")
	corpus, err := har.InitializeRoutes(routes)
	require.NoError(t, err)
	for _, route := range corpus {
		fmt.Printf("%v %v\n", route.Method, route.Path)
	}

}
