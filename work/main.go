// Command work starts the chainr worker.
//
// The worker processes pending jobs, manages dependencies, runs jobs on
// Kubernetes and update status and events.
package main

import (
	"log"

	"github.com/Tyrame/chainr/work/internal/worker"
)

func main() {
	log.Println("Starting chainr worker")

	w := worker.New()
	log.Fatal(w.Start())
}
