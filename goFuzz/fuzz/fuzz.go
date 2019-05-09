package fuzz

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/mutator"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/pkg/errors"
)

// Fuzz starts the fuzzer
func Fuzz(corpus []*http.Request, client *httpclient.Client) {
	// Parse routes
	routes := route.LoadRoutes()
	mutator := mutator.New(corpus, routes)
	for {
		// Get next request
		request := mutator.Next()
		// No routes left to explore
		if request == nil {
			break
		}

		// Send it.
		fmt.Printf("%v %v\n", request.Method, request.URL)
		resp, err := client.Do(request)
		if err != nil {
			err := errors.Wrap(err, "")
			log.Fatalf("%+v\n", err)
		}
		// Collect our deltas
		mutator.UpdateCoverage(resp)
		fmt.Println("Breaking!!!!")
		break
	}
}
