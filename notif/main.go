// Command notif starts the chainr notifier.
//
// The notifier analyzes events, and sends notifications on supported medias.
package main

import (
	"log"

	"github.com/Tyrame/chainr/notif/internal/worker"
)

func main() {
	log.Println("Starting chainr notifier")

	w := worker.New()
	w.Start()
}
