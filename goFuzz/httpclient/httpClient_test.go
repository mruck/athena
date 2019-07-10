package httpclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func urlFromTestServer(t *testing.T, ts *httptest.Server) *url.URL {
	url, err := url.Parse(ts.URL)
	require.NoError(t, err)
	return url
}

func TestConnect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()
	url := urlFromTestServer(t, ts)

	request, err := http.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	requests := []*http.Request{request}

	client, err := New(url)
	require.NoError(t, err)

	err = client.DoAll(requests)
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
	url := urlFromTestServer(t, ts)

	route := fmt.Sprintf("%s/getCookie", url)
	request1, err := http.NewRequest("GET", route, nil)
	require.NoError(t, err)

	route = fmt.Sprintf("%s/checkCookie", url)
	request2, err := http.NewRequest("GET", route, nil)
	require.NoError(t, err)

	requests := []*http.Request{request1, request2}

	client, err := New(url)
	require.NoError(t, err)

	err = client.DoAll(requests)
	require.NoError(t, err)
}

func TestHealthCheck(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == healthCheckRoute {
			// TODO: rails endpoint returns 404 so emulate that, fix rails to return 200
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500 - Something bad happened!"))
		require.NoError(t, err)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()
	url := urlFromTestServer(t, ts)

	client, err := New(url)
	require.NoError(t, err)

	alive, err := client.HealthCheck()
	require.NoError(t, err)
	require.Equal(t, true, alive)
}
