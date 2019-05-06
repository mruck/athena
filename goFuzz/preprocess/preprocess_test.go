package preprocess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFromHar(t *testing.T) {
	_, err := fileToHar("test/login_har.json")
	require.NoError(t, err)
}

func TestHarToRequest(t *testing.T) {
	_, err := harToRequest(nil)
	require.NoError(t, err)
}
