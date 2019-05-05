package main

import (
	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
)

func main() {
	// Retrieve HTTP state for logging in
	httpstate := preprocess.GetHTTPState()
	client := httpclient.New(httpstate)
	// Parse initial corpus
	corpus := preprocess.GetCorpus()
	fuzz.Launch(corpus, client)

}
