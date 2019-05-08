package preprocess

import (
	"testing"

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
