package fuzz

import (
	"net/http"
)

func start() {
	// init logger
	// health check target
	// dump pluralizations
	// connect to target db
}

// Launch fuzzer
func Launch(corpus []*http.Request, client *http.Client) {
	start()
	//fuzz stats for benchmarking?
}
