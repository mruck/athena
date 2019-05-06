package fuzz

import (
	"github.com/mruck/athena/goFuzz/httpclient"
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
func Launch(corpus *preprocess.Corpus, client *httpclient.HTTPClient) {
	start()
	//fuzz stats for benchmarking?
}
