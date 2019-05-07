package main

import (
	"log"

	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
)

// TODO: this should be in the shared mount.  Not sure a way around hard
// coding this
const harPath = "tests/login_har.json"

func main() {
	// Retrieve HTTP state for logging in
	login, err := preprocess.GetLogin(harPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	client, err := httpclient.New(login)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	// Parse initial corpus
	corpus := preprocess.GetCorpus()
	fuzz.Launch(corpus, client)

}
