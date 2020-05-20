package worker

import "testing"

func TestNewInfo(t *testing.T) {
	info := NewInfo()
	if len(info.Name) == 0 {
		t.Errorf("info.Name is empty, expected non-empty")
	}
	if info.Queue != "runs:work" {
		t.Errorf("info.Queue = %v, expected runs:work", info.Queue)
	}
	if info.ProcessQueue != "runs:worker:"+info.Name {
		t.Errorf("info.ProcessQueue = %v, expected runs:worker:%v", info.ProcessQueue, info.Name)
	}
}
