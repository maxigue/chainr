package worker

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"errors"

	"github.com/Tyrame/chainr/notif/internal/notifier"
)

// As the worker mostly works as a black box,
// most tests in this file are written with stubs.
// To embed inline documentation on the behaviour,
// behaviour-driven testing is still present, despite
// not being perfectly suited.

var testInfo = Info{"xyz", "events:notif", "events:notifier:xyz"}

type notifierStub struct{}

func (n notifierStub) Dispatch(event notifier.Event) error {
	return nil
}

type brokenNotifierStub struct{}

func (n brokenNotifierStub) Dispatch(event notifier.Event) error {
	return errors.New("failed")
}

type eventStoreStub struct{}

func (es eventStoreStub) NextEvent() (string, error) {
	return "event:abc", nil
}

func (es eventStoreStub) GetEvent(eventID string) (notifier.Event, error) {
	return notifier.Event{
		Type:    "SUCCESS",
		Title:   "Event title",
		Message: "Event message",
	}, nil
}

func (es eventStoreStub) Close(eventId string) error {
	return nil
}

type recyclerStub struct{}

func (r recyclerStub) StartSync() {}

func TestDispatchNextEvent(t *testing.T) {
	Convey("Scenario: dispatch a valid event", t, func() {
		Convey("Given an event is retrieved", func() {
			Convey("When the dispatch is successful", func() {
				w := Worker{&eventStoreStub{}, &notifierStub{}, &recyclerStub{}}

				Convey("The loop should continue without error", func() {
					err := w.DispatchNextEvent()
					So(err, ShouldBeNil)
				})
			})

			Convey("When the dispatch fails", func() {
				w := Worker{&eventStoreStub{}, &brokenNotifierStub{}, &recyclerStub{}}

				Convey("The loop should still continue without error", func() {
					err := w.DispatchNextEvent()
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
