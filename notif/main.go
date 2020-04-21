// Command notif starts the chainr notifier.
//
// The notifier analyzes events, and sends notifications on supported medias.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

var port = 8080

func init() {
	val, ok := os.LookupEnv("PORT")
	if ok {
		p, err := strconv.Atoi(val)
		if err != nil {
			panic(err.Error())
		}
		port = p
	}
}

func main() {
	log.Println("Starting chainr notifier")

	// TODO: this block is only for bootstrapping
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	})

	addr := fmt.Sprintf(":%d", port)
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
