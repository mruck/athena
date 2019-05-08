package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
	"github.com/mruck/athena/goFuzz/util"
)

// TODO: this should be in the shared mount.
// Add to target img?
const harPath = "preprocess/test/login_har.json"

func must(check bool, format string, args ...interface{}) {
	if !check {
		log.Fatalf(format, args...)
	}
}

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

	// Parse the URL first.
	url, err := url.Parse(fmt.Sprintf("http://%s:%s", host, port))
	must(err == nil, "%+v", err)

	// Get a new client.
	client, err := httpclient.New(url)
	must(err == nil, "%+v", err)

	// Health check
	alive, err := client.HealthCheck()
	must(err == nil, "%+v", err)
	must(alive, "target app not alive")

	// Send login har.
	err = client.DoAll(login)
	must(err == nil, "%+v", err)

	fuzz.Fuzz(corpus, client)
}
