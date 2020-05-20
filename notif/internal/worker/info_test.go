package worker

import "testing"

func TestNewInfo(t *testing.T) {
	info := NewInfo()
	if len(info.Name) == 0 {
		t.Errorf("info.Name is empty, expected non-empty")
	}
	if info.Queue != "events:notif" {
		t.Errorf("info.Queue = %v, expected events:notif", info.Queue)
	}
	if info.ProcessQueue != "events:notifier:"+info.Name {
		t.Errorf("info.ProcessQueue = %v, expected events:notifier:%v", info.ProcessQueue, info.Name)
	}
}
