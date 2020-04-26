// Command work starts the chainr worker.
//
// The worker processes pending jobs, manages dependencies, runs jobs on
// Kubernetes and update status and events.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	log.Println("Starting chainr worker")

	port := 8080
	val, ok := os.LookupEnv("PORT")
	if ok {
		p, err := strconv.Atoi(val)
		if err != nil {
			panic(err.Error())
		}
		port = p
	}

	// TODO: this block is only for bootstrapping
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	})

	addr := fmt.Sprintf(":%d", port)
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
