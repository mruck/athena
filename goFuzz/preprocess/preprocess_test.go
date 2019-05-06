package preprocess

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHarToRequest(t *testing.T) {
	har, err := unmarshalHar("test/login_har.json")
	require.NoError(t, err)
	requests, err := harToRequest(har)
	require.NoError(t, err)
	fmt.Println(requests)
}