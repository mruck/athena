package param

import (
	"testing"

	"github.com/go-openapi/spec"
	"github.com/mruck/athena/lib/util"
	"github.com/stretchr/testify/require"
)

func TestMutate(t *testing.T) {
	// Read in a spec.Parameter
	var param spec.Parameter
	err := util.UnmarshalFile("param.test", &param)
	require.NoError(t, err)
	// Create an obj for tracking state
	state := New(param)
	// Mutate
	state.MockData()
	// Check
	casted, ok := state.Next.(map[string]interface{})
	require.True(t, ok)
	_, ok = casted["color"]
	require.True(t, ok)
	_, ok = casted["name"]
	require.True(t, ok)
	_, ok = casted["text_color"]
	require.True(t, ok)
}

func TestGetPathParams(t *testing.T) {
	path := "/users/{username}/preferences/avatar/pick/{id}"
	paramUser := New(*spec.PathParam("username"))
	paramID := New(*spec.PathParam("id"))

	params := getPathParams(path)
	require.Equal(t, paramUser, params[0])
	require.Equal(t, paramID, params[1])
}
