package httpclient

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

// HTTPClient contains a connection
type HTTPClient struct {
	Client *http.Client
}

// New initializes a HTTPClient using the state provided
func New(requests []*http.Request) (*HTTPClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar: jar,
	}
	for _, request := range requests {
		resp, err := client.Do(request)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 200 {
			err = fmt.Errorf("status code: %v", resp.StatusCode)
			return nil, err
		}
		// Read cookies
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			fmt.Printf("cookie: %v\n", cookie.String())
		}
	}
	httpclient := &HTTPClient{Client: client}
	return httpclient, nil
}
