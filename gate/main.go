// Command gate starts the chainr API Gateway.
//
// The gate is the entry point of chainr cloud services. It provides a coherent
// api on top of micro-services, and adds hypermedia informations on top of
// responses.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var configFile string = "config.yaml"

func init() {
	val, ok := os.LookupEnv("CONFIG_FILE")
	if ok {
		configFile = val
	}
}

func main() {
	log.Println("Starting chainr gate")
	cfg, err := LoadConfig(configFile)
	if err != nil {
		log.Println("Configuration loading failed:", err.Error())
		log.Println("Using default configuration")
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Println("Listening on", addr)
	http.ListenAndServe(addr, NewHandler(cfg))
}
