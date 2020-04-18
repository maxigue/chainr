package run

import (
	"testing"

	"net/http"
)

func TestNewRunner(t *testing.T) {
	body := `{"kind": "Pipeline"}`
	_, err := newRunner([]byte(body))
	if err != nil {
		t.Errorf("newRunner on a valid pipeline returned nil")
	}
}

func TestNewRunnerBadFormat(t *testing.T) {
	_, err := newRunner([]byte("{invalid}"))
	if err == nil {
		t.Fatal("newRunner on an invalid format returned a nil error")
	}
	if err.Status() != http.StatusBadRequest {
		t.Errorf("err.Status() = %v, expected %v", err.Status(), http.StatusBadRequest)
	}
}

func TestNewRunnerBadKind(t *testing.T) {
	body := `{"kind": "Unknown"}`
	_, err := newRunner([]byte(body))
	if err == nil {
		t.Fatal("newRunner on an invalid kind returned a nil error")
	}
	if err.Status() != http.StatusBadRequest {
		t.Errorf("err.Status() = %v, expected %v", err.Status(), http.StatusBadRequest)
	}
}
