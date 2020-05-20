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
	es       EventStore
	n        notifier.Notifier
	recycler Recycler
}

type EventStore interface {
	// Actively listens to new events, and when a new event is available,
	// returns an arbitrary string identifier referencing it.
	// This identifier is used to refer to the event in store operations.
	// Events returned by this function must be closed when no longer used.
	NextEvent() (string, error)

	// Returns a structure describing the event referenced by the
	// arbitrary identifier.
	GetEvent(eventID string) (notifier.Event, error)

	// Closes the event corresponding to the identifier.
	// Post-dispatch operations are done in this function.
	Close(eventID string) error
}

type Recycler interface {
	// Start synchronizing with the recycler.
	// As the synchronization is a loop, it must be called in a goroutine.
	StartSync()
}

func New() Worker {
	info := NewInfo()
	return Worker{
		NewRedisEventStore(info),
		notifier.NewLogNotifier(),
		NewRecycler(info),
	}
}

// Start launches the worker loop.
// It runs indefinitely.
// Upon starting, it synchronizes with the recycler.
func (w Worker) Start() {
	go w.recycler.StartSync()

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
	eventID, err := w.es.NextEvent()
	if err != nil {
		return err
	}

	event, err := w.es.GetEvent(eventID)
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

	if err := w.es.Close(eventID); err != nil {
		log.Printf("Unable to close event %v: %v", eventID, err.Error())
	}

	return nil
}
