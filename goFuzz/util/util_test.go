package util

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatchHostPort(t *testing.T) {
	req, err := http.NewRequest("GET", "http://192.168.3.8:8080", nil)
	require.NoError(t, err)
	PatchHostPort(req, "localhost", "1234")
	require.Equal(t, "localhost:1234", req.URL.Host)
}
