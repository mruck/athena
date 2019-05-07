package fuzz

import (
	"net/http"

	"github.com/mruck/athena/goFuzz/httpclient"
)

func start() {
	// init logger
	// health check target
	// dump pluralizations
	// connect to target db
}

// Launch fuzzer
func Launch(corpus []*http.Request, client *httpclient.Client) {
	start()
	//fuzz stats for benchmarking?
}
