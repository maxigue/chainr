package pipeline

import "testing"

func TestNew(t *testing.T) {
	p := New()
	if p.Kind != "Pipeline" {
		t.Errorf("p.Kind = %v, expected Pipeline", p.Kind)
	}
}
