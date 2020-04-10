package pipeline

import (
	"testing"

	"github.com/Tyrame/chainr/sched/config"
)

func TestHandler(t *testing.T) {
	h := NewHandler(config.Configuration{})
	if h == nil {
		t.Fatal("h is nil, expected value")
	}
}
