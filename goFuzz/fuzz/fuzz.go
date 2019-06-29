package fuzz

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mruck/athena/goFuzz/httpclient"
	"github.com/mruck/athena/goFuzz/mutator"
	"github.com/mruck/athena/goFuzz/route"
	"github.com/mruck/athena/lib/log"
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

	fmt.Printf("Code Counts: %s\n", string(codes))
	fmt.Printf("Final Coverage: %v\n", mutator.SrcCoverage.Cumulative)
	fmt.Printf("Success Ratio: %v\n", successRatio)
	fmt.Printf("Total Requests: %v\n", totalRequests)
	// TODO: figure out how to parse this in bash because right now messes up sanity.sh
	//log.Infof("Code Counts: %s", string(codes))
	//log.Infof("Final Coverage: %v\n", mutator.Coverage.Cumulative)
	//log.Infof("Success Ratio: %v\n", successRatio)
	//log.Infof("Total Requests: %v\n", totalRequests)

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
		resp, err := client.Do(request)
		if err != nil {
			log.Error(err)
		}

		// Collect our deltas
		err = mutator.UpdateState(resp)
		if err != nil {
			// Log the error with some additional context
			mutator.LogError(err)
		}
	}

	logStats(client, mutator)
}
