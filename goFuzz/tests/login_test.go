package main

import (
	"net/http"
	"testing"

	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
	"github.com/mruck/athena/goFuzz/util"
	"github.com/stretchr/testify/require"
)

// TestLogin tests that we can login to discourse from a HAR file
func TestLogin(t *testing.T) {
	harPath := "../preprocess/test/login_har.json"
	host := "localhost"
	port := "8080"
	// Retrieve HTTP state for logging in
	login, err := preprocess.GetLogin(harPath)
	require.NoError(t, err)
	util.PatchRequestsHostPort(login, host, port)
	// Create a client that logs in
	client, err := httpclient.New(login)
	require.NoError(t, err)
	require.NotNil(t, client)
	// Ping a route that we have access to only if we are logged in
	req, err := http.NewRequest("GET", "http://localhost:8080/admin", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}