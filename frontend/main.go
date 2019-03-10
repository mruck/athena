package main

//Listen on a port
//Read in spec from body
//Dump original spec to disc
//Shell out and spin this up

import (
	"log"
	"net/http"
)

func main() {
	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
