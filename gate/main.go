// Command gate starts the chainr API Gateway.
//
// The gate is the entry point of chainr cloud services. It provides a coherent
// api on top of micro-services, and adds hypermedia informations on top of
// responses.
package main

import (
	"fmt"
	"log"
)

func main() {
	log.Println("Starting gate...")

	cfg, err := LoadConfig("config.yaml")
	if (err != nil) {
		log.Fatal(err.Error())
	}
	fmt.Println("Port:", cfg.Port)

	// This makes the command wait indefinitely
	for {
	}
}
