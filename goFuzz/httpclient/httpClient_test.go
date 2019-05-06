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

	client, err := New(requests)
	require.NoError(t, err)
	fmt.Println(client)
}

func TestCookieManagement(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cookie := &http.Cookie{Name: "TestCookieName", Value: "TestCookieValue", MaxAge: 3600}
		if req.URL.Path == "/getCookie" {
			fmt.Printf("setting cookie\n")
			http.SetCookie(w, cookie)
			// Respond with new cookie
			return
		}
		// Check cookie
		newCookie, err := req.Cookie(cookie.Name)
		require.NoError(t, err)
		fmt.Printf("%v %v\n", newCookie.Value, cookie.Value)
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

	client, err := New(requests)
	require.NoError(t, err)
	fmt.Println(client)
}
