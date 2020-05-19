// Command recycle starts the chainr recycler.
//
// The recycler collects items that were not fully processed by workers
// (e.g. due to outages), and re-schedules them.
package main

import (
	"log"
	"time"
)

func main() {
	log.Println("Starting chainr recycler")

	for {
		time.Sleep(1 * time.Second)
	}
}
