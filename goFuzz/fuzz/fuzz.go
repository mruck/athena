package fuzz

import (
	"fmt"
	"net/http"

	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/mutator"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/util"
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
		util.Must(err == nil, "%+v", errors.WithStack(err))

		// Collect our deltas
		err = mutator.UpdateCoverage(resp)
		util.Must(err == nil, "%+v", err)
	}
}
