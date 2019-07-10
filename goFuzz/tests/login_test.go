package main

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
	"github.com/stretchr/testify/require"
)

// TestLogin tests that we can login to discourse from a HAR file
func TestLogin(t *testing.T) {
	t.Skip()
	harPath := "../tests/login_har.json"
	// Retrieve HTTP state for logging in
	login, err := preprocess.GetLogin(harPath)
	require.NoError(t, err)

	url, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	// Create a client.
	client, err := httpclient.New(url)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Healthcheck first
	ok, err := client.HealthCheck()
	require.NoError(t, err)
	require.True(t, ok)

	// Log in with this client.
	err = client.DoAll(login)
	require.NoError(t, err)

	// Ping a route that we have access to only if we are logged in
	req, err := http.NewRequest("GET", "http://localhost:8080/admin", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
