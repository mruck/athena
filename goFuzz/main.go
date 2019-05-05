package main

import (
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
)

func main() {
	// Retrieve HTTP state for logging in
	HTTPState := preprocess.GetHTTPState()
	HTTPClient := httpclient.New(HTTPState)

}
