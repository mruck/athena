package main

import (
	"fmt"
	"log"

	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
	"github.com/mruck/athena/goFuzz/util"
)

// TODO: this should be in the shared mount.
// Add to target img?
const harPath = "preprocess/test/login_har.json"

func main() {
	port := util.MustGetTargetAppPort()
	host := util.MustGetTargetAppHost()
	// Parse initial corpus
	corpus := preprocess.GetCorpus()
	// TODO: Patch host/port in corpus
	// Retrieve HTTP state for logging in
	login, err := preprocess.GetLogin(harPath)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	util.PatchRequestsHostPort(login, host, port)
	// Health check
	url := fmt.Sprintf("http://%v:%v", host, port)
	alive, err := httpclient.HealthCheck(url)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if !alive {
		log.Fatal("Target app not alive")
	}
	// Login to target app
	client, err := httpclient.New(login)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	fuzz.Launch(corpus, client)
}
