package pipeline

import (
	"testing"

	"net/http"
)

func TestNew(t *testing.T) {
	p := New()
	if p.Kind != "Pipeline" {
		t.Errorf("p.Kind = %v, expected Pipeline", p.Kind)
	}
	if p.Jobs == nil {
		t.Errorf("p.Jobs is nil, expected map")
	}
}

func TestNewFromSpec(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"jobs": {
			"job1": {
				"image": "busybox",
				"run": "exit 0"
			}
		}
	}`)

	p, err := NewFromSpec(spec)
	if err != nil {
		t.Fatal("err = nil, expected not nil")
	}
	if image := p.Jobs["job1"].Image; image != "busybox" {
		t.Errorf("image = %v, expected busybox", image)
	}
}

func TestNewFromSpecBadFormat(t *testing.T) {
	spec := []byte(`{invalid}`)
	_, err := NewFromSpec(spec)
	if err == nil {
		t.Fatal("NewFromSpec from an invalid format returned a nil error")
	}
	if err.Status() != http.StatusBadRequest {
		t.Errorf("err.Status() = %v, expected %v", err.Status(), http.StatusBadRequest)
	}
}

func TestNewFromSpecBadSchema(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"invalid": "hello"
	}`)
	_, err := NewFromSpec(spec)
	if err == nil {
		t.Fatal("NewFromSpec from an invalid schema returned a nil error")
	}
	if err.Status() != http.StatusBadRequest {
		t.Errorf("err.Status() = %v, expected %v", err.Status(), http.StatusBadRequest)
	}
}

func TestNewFromSpecBadDeps(t *testing.T) {
	// TODO: implement
}

func TestValidate(t *testing.T) {
	// TODO: implement
}
