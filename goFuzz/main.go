package main

import (
	"fmt"
	"log"

	"github.com/mruck/athena/goFuzz/fuzz"
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/preprocess"
)

func main() {
	// Retrieve HTTP state for logging in
	login := preprocess.GetLogin()
	client, err := httpclient.New(login)
	if err != nil {
		err = fmt.Errorf("failed to create http client: %v", err)
		log.Fatal(err)
	}
	// Parse initial corpus
	corpus := preprocess.GetCorpus()
	fuzz.Launch(corpus, client)

}
