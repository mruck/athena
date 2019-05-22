package main

//Listen on a port
//Read in spec from body
//Dump original spec to disc
//Shell out and spin this up

import (
	"net/http"

	"github.com/mruck/athena/frontend/server"
	"github.com/mruck/athena/lib/log"
)

func main() {
	router, err := server.NewRouter()
	if err != nil {
		panic(err)
	}

	log.Infof("Listening on 8081\n")
	log.Fatalf("%+v", http.ListenAndServe(":8081", router))
}
