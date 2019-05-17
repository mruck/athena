package route

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const bodyTestfile = "test/body.json"
const pathTestfile = "test/path.json"
const queryTestfile = "test/query.json"

func TestSetPathParam(t *testing.T) {
	routes := FromSwagger(pathTestfile)
	routes[0].Mutate()
	path := routes[0].SetPathParams()
	// Make sure it has the path
	require.True(t, strings.Contains(path, "/pet"))
	// Make sure we replace {}
	require.False(t, strings.Contains(path, "{"))
	// TODO: do a check via a regex!
}

func TestGetQueryStr(t *testing.T) {
	routes := FromSwagger(queryTestfile)
	routes[0].Mutate()
	query := routes[0].GetQueryStr()
	// TODO: do a check via a regex!
	require.True(t, strings.Contains(query, "username"))
}

func TestGetBodyParams(t *testing.T) {
	routes := FromSwagger(bodyTestfile)
	routes[0].Mutate()
	bodyReader, err := routes[0].GetBodyParams()
	require.NoError(t, err)
	require.NotNil(t, bodyReader)
}

func TestSendQueryParams(t *testing.T) {
}

func TestSendBodyParams(t *testing.T) {
}

func TestSendPathParams(t *testing.T) {
}
