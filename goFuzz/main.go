package main

import (
	"log"

	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
)

func main() {
	// Retrieve HTTP state for logging in
	login, err := preprocess.GetLogin()
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
