package main

import (
	"fmt"
	"log"
	"net/http"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprint(w, "Hello world")
	if err != nil {
		panic(err)
	}
}

func main() {
	// Registers HandleRoot func to handle requests to "/"
	http.HandleFunc("/", HandleRoot)
	// Starts the HTTP server on port 8080 and runs until it gets an error 
	log.Fatal(http.ListenAndServe(":8080", nil))
}
