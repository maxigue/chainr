package notifier

import "testing"

func TestDispatch(t *testing.T) {
	n := NewLogNotifier()

	successfulEvents := []Event{
		Event{"START", "title", "message"},
		Event{"SUCCESS", "title", "message"},
		Event{"FAILURE", "title", "message"},
		Event{"UNKNOWN", "title", "message"},
		Event{},
	}

	for _, event := range successfulEvents {
		err := n.Dispatch(event)
		if err != nil {
			t.Errorf("err = %v, expected nil", err)
		}
	}
}
