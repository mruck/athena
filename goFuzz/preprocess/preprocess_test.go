package preprocess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHarToRequest(t *testing.T) {
	har, err := unmarshalHar("test/login_har.json")
	require.NoError(t, err)
	requests, err := har.toRequests()
	require.NoError(t, err)
	require.NotNil(t, requests)
	// TODO: check headers?
	// TODO: check body?
}
