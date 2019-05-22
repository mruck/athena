package fuzz

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/mutator"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/lib/util"
	"github.com/pkg/errors"
)

func logStats(client *httpclient.Client, mutator *mutator.Mutator) {
	totalRequests := 0
	for _, num := range client.StatusCodes {
		totalRequests += num
	}

	successRatio := float64(client.StatusCodes[200]) / float64(totalRequests)

	stringified := make(map[string]int, len(client.StatusCodes))
	for k, v := range client.StatusCodes {
		stringified[strconv.Itoa(k)] = v
	}
	codes, err := json.Marshal(stringified)
	util.Must(err == nil, "%+v\n", errors.WithStack(err))

	fmt.Println("Code Counts: ", string(codes))
	fmt.Printf("Final Coverage: %v\n", mutator.Coverage.Cumulative)
	fmt.Printf("Success Ratio: %v\n", successRatio)
	fmt.Printf("Total Requests: %v\n", totalRequests)

}

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
	logStats(client, mutator)
}
