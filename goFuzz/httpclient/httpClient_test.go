package httpclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	request, err := http.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	requests := []*http.Request{request}

	_, err = New(requests)
	require.NoError(t, err)
}

func TestCookieManagement(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cookie := &http.Cookie{Name: "TestCookieName", Value: "TestCookieValue", MaxAge: 3600}
		if req.URL.Path == "/getCookie" {
			// Respond with new cookie
			http.SetCookie(w, cookie)
			return
		}
		// Check cookie
		newCookie, err := req.Cookie(cookie.Name)
		require.NoError(t, err)
		require.Equal(t, newCookie.Value, cookie.Value)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	url := fmt.Sprintf("%s/getCookie", ts.URL)
	request1, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	url = fmt.Sprintf("%s/checkCookie", ts.URL)
	request2, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	requests := []*http.Request{request1, request2}

	_, err = New(requests)
	require.NoError(t, err)
}

func TestHealthCheck(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == healthCheckRoute {
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500 - Something bad happened!"))
		require.NoError(t, err)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()
	alive, err := HealthCheck(ts.URL)
	require.NoError(t, err)
	require.Equal(t, true, alive)
}
