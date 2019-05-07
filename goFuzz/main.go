package main

import (
	"log"

	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
	"github.com/mruck/athena/goFuzz/util"
)

// TODO: this should be in the shared mount.  Not sure a way around hard
// coding this
const harPath = "tests/login_har.json"
const host = "localhost"
const port = "8080"

func main() {
	// Retrieve HTTP state for logging in
	login, err := preprocess.GetLogin(harPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	util.PatchRequestsHostPort(login, host, port)
	client, err := httpclient.New(login)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	// Parse initial corpus
	corpus := preprocess.GetCorpus()
	fuzz.Launch(corpus, client)

}
