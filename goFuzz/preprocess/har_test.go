package preprocess

import (
	"fmt"
	"testing"

	"github.com/mruck/athena/goFuzz/route"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalHar(t *testing.T) {
	har := unmarshalHar("test/login_har.json")
	// Pick something random to check for equality
	request0 := har.Log.Entries[0].Request
	require.Equal(t, "http://localhost:50121/session/csrf?_=1548444062137", request0.URL)
	//util.PrettyPrint(har)
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
