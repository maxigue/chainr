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
)

func main() {
	log.Println("Starting chainr gate")
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err.Error())
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Println("Listening on", addr)
	http.ListenAndServe(addr, nil)
}
