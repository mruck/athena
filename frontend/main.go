package main

//Listen on a port
//Read in spec from body
//Dump original spec to disc
//Shell out and spin this up

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	router := NewRouter()

	fmt.Printf("Listening on 1111\n")
	log.Fatal(http.ListenAndServe(":1111", router))
}
