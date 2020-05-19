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

type notifierStub struct{}

func (n notifierStub) Dispatch(event notifier.Event) error {
	return nil
}

type brokenNotifierStub struct{}

func (n brokenNotifierStub) Dispatch(event notifier.Event) error {
	return errors.New("failed")
}

type eventStoreStub struct{}

func (es eventStoreStub) NextEvent() (notifier.Event, error) {
	return notifier.Event{
		Type:    "SUCCESS",
		Title:   "Event title",
		Message: "Event message",
	}, nil
}

func TestDispatchNextEvent(t *testing.T) {
	Convey("Scenario: dispatch a valid event", t, func() {
		Convey("Given an event is retrieved", func() {
			Convey("When the dispatch is successful", func() {
				w := Worker{&eventStoreStub{}, &notifierStub{}}

				Convey("The loop should continue without error", func() {
					err := w.DispatchNextEvent()
					So(err, ShouldBeNil)
				})
			})

			Convey("When the dispatch fails", func() {
				w := Worker{&eventStoreStub{}, &brokenNotifierStub{}}

				Convey("The loop should still continue without error", func() {
					err := w.DispatchNextEvent()
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
