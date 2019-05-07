package util

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestPatchHostPort(t *testing.T) {
	req, err := http.NewRequest("GET", "http://192.168.3.8:8080", nil)
	require.NoError(t, err)
	PatchRequestHostPort(req, "localhost", "1234")
	require.Equal(t, "localhost:1234", req.URL.Host)
}

func TestRequestsPatchHostPort(t *testing.T) {
	req0, err := http.NewRequest("GET", "http://192.168.3.8:8080", nil)
	require.NoError(t, err)
	req1, err := http.NewRequest("GET", "http://192.168.3.8:8080", nil)
	require.NoError(t, err)
	requests := []*http.Request{req0, req1}
	PatchRequestsHostPort(requests, "localhost", "1234")
	require.Equal(t, "localhost:1234", req0.URL.Host)
	require.Equal(t, "localhost:1234", req1.URL.Host)
	require.Equal(t, "localhost:1234", req0.Host)
	require.Equal(t, "localhost:1234", req1.Host)
}
