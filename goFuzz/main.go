package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/util"
)

// TODO: this should be in the shared mount.
// Add to target img?
const harPath = "preprocess/test/login_har.json"
const swaggerPath = "swagger.json"

func main() {
	port := util.MustGetTargetAppPort()
	host := util.MustGetTargetAppHost()

	// Load swagger info
	routes, err := route.LoadRoutes(swaggerPath)
	util.Must(err == nil, "%+v", err)

	// Parse initial corpus
	corpus := preprocess.GetCorpus()

	// Retrieve HTTP state for logging in
	login, err := preprocess.GetLogin(harPath)
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

	fuzz.Fuzz(corpus, client)
}
