// Package worker contains tools to interact with
// the events store, and start events processing.
// A worker consumes events, and dispatches them through notifiers.
package worker

import (
	"log"
	"time"

	"github.com/Tyrame/chainr/notif/internal/notifier"
)

type Worker struct {
	es EventStore
	n  notifier.Notifier
}

type EventStore interface {
	// Actively listens to new events, and when a new event is available,
	// returns a structure describing it.
	NextEvent() (notifier.Event, error)
}

func New() Worker {
	return Worker{
		NewRedisEventStore(),
		notifier.NewLogNotifier(),
	}
}

// Start launches the worker loop.
// It runs indefinitely.
func (w Worker) Start() {
	for {
		if err := w.DispatchNextEvent(); err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
		}
	}
}

// DispatchNextEvent is a blocking function, listening for a new event,
// and dispatching it.
func (w Worker) DispatchNextEvent() error {
	event, err := w.es.NextEvent()
	if err != nil {
		return err
	}

	log.Printf(`Dispatching event
	type: %v
	title: %v
	message: %v`, event.Type, event.Title, event.Message)

	if err := w.n.Dispatch(event); err != nil {
		log.Println("Error while dispatching event:", err)
	}

	return nil
}
