// Command notif starts the chainr notifier.
//
// TODO
// The notifier allows to run jobs on the Kubernetes cluster it is deployed on.
// It manages the whole dependency tree, and checks all preconditions are met
// before running a job.
// It exposes an API allowing to run a pipeline, and get its execution status.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Tyrame/chainr/notif/internal/config"
)

var configFile string = "config.yaml"

func init() {
	val, ok := os.LookupEnv("CONFIG_FILE")
	if ok {
		configFile = val
	}
}

func main() {
	log.Println("Starting chainr notifier")
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Println("Configuration loading failed:", err.Error())
		log.Println("Using default configuration")
	}

	// TODO: this block is only for bootstrapping
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
