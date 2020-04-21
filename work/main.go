// Command work starts the chainr worker.
//
// TODO
// The worker allows to run jobs on the Kubernetes cluster it is deployed on.
// It manages the whole dependency tree, and checks all preconditions are met
// before running a job.
// It exposes an API allowing to run a pipeline, and get its execution status.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Tyrame/chainr/work/internal/config"
	"github.com/Tyrame/chainr/work/internal/k8s"
)

var configFile string = "config.yaml"

func init() {
	val, ok := os.LookupEnv("CONFIG_FILE")
	if ok {
		configFile = val
	}
}

func main() {
	log.Println("Starting chainr worker")
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Println("Configuration loading failed:", err.Error())
		log.Println("Using default configuration")
	}

	k8s, err := k8s.New(cfg.Kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// TODO: this block is only for bootstrapping
	_ = k8s
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
