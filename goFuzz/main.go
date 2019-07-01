package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/lib/exception"
	"github.com/mruck/athena/lib/util"
)

// TODO: this should be in the shared mount.
// Add to target img?
const harLogin = "tests/login_har.json"
const swaggerPath = "tests/discourseSwagger.json"
const harCorpus = "tests/corpus_har.json"

func readDB() {
	_, found := os.LookupEnv("READ_DB")
	if !found {
		return
	}
	// Read database and exit
	exception.ReadDB()
	os.Exit(0)
}

func main() {
	// Read the database and exit
	readDB()

	port := util.MustGetTargetAppPort()
	host := util.MustGetTargetAppHost()

	// Load swagger info
	routes := route.FromSwagger(swaggerPath)

	// Parse initial corpus
	corpus := preprocess.GetCorpus(routes, harCorpus)

	// Retrieve HTTP state for logging in
	login, err := preprocess.GetLogin(harLogin)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// Parse the URL
	url, err := url.Parse(fmt.Sprintf("http://%s:%s", host, port))
	util.Must(err == nil, "%+v", err)

	// Get a new client.
	client, err := httpclient.New(url)
	util.Must(err == nil, "%+v", err)

	// Health check
	alive, err := client.HealthCheck()
	util.Must(err == nil, "%+v", err)
	util.Must(alive, "target app not alive")

	// Login
	err = client.DoAll(login)
	util.Must(err == nil, "%+v", err)

	fuzz.Fuzz(client, routes, corpus)
}
