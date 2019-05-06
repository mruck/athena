package fuzz

import (
	"net/http"

	"github.com/mruck/athena/goFuzz/preprocess"
)

func start() {
	// snapshotting?
	// init logger
	// health check target
	// dump pluralizations
	// connect to target db
}

// Launch fuzzer
func Launch(corpus *preprocess.Corpus, client *http.Client) {
	start()
	//fuzz stats for benchmarking?
}
