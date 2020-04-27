// Command sched starts the chainr scheduler.
//
// The scheduler allows to schedule jobs, for workers to run on the Kubernetes
// cluster.
// It exposes an API allowing to run a pipeline, and get its execution status.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func addr() string {
	port := 8080
	val, ok := os.LookupEnv("PORT")
	if ok {
		p, err := strconv.Atoi(val)
		if err != nil {
			panic(err.Error())
		}
		port = p
	}

	return fmt.Sprintf(":%d", port)
}

func main() {
	log.Println("Starting chainr scheduler")

	addr := addr()
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, NewHandler()))
}
