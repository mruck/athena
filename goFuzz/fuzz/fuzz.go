package fuzz

import (
	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/mutator"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/goFuzz/util"
	"github.com/pkg/errors"
)

// Fuzz starts the fuzzer
func Fuzz(client *httpclient.Client, routes []*route.Route, corpus []*route.Route) {
	// Parse routes
	mutator := mutator.New(routes, corpus)
	for {
		// Get next request
		request := mutator.Next()
		// No routes left to explore
		if request == nil {
			break
		}

		// Send it.
		httpclient.PrettyPrintRequest(request)
		resp, err := client.Do(request)
		util.Must(err == nil, "%+v", errors.WithStack(err))

		// Collect our deltas
		err = mutator.UpdateState(resp)
		util.Must(err == nil, "%+v", err)
	}
	util.PrettyPrintStruct(client.StatusCodes)
}
