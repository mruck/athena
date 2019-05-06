package preprocess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalHar(t *testing.T) {
	har, err := unmarshalHar("test/login_har.json")
	require.NoError(t, err)
	// Pick something random to check for equality
	request0 := har.Log.Entries[0].Request
	require.Equal(t, request0.URL, "http://localhost:50121/login")
	//util.PrettyPrint(har)
}