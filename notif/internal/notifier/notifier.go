// Package notifier contains notifiers on different medias.
// Notifiers dispatch events.
package notifier

type Notifier interface {
	// Runs the job on the cloud provider.
	// Blocks until the job completes.
	Dispatch(event Event) error
}

// Type can be:
// - START
// - SUCCESS
// - FAILURE
type Event struct {
	Type    string
	Title   string
	Message string
}
