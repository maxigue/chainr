package run

import (
	"testing"

	"net/http"
)

func TestNewPipelineRunner(t *testing.T) {
	spec := `{
		"kind": "Pipeline",
		"jobs": {}
	}`
	_, err := newPipelineRunner([]byte(spec))
	if err != nil {
		t.Errorf("newPipelineRunner on a valid pipeline returned nil")
	}
}

func TestNewPipelineRunnerBadFormat(t *testing.T) {
	spec := `{
		"kind": "Pipeline",
		"test": invalid
	}`
	_, err := newPipelineRunner([]byte(spec))
	if err == nil {
		t.Fatal("newPipelineRunner on an invalid pipeline returned a nil error")
	}
	if err.Status() != http.StatusBadRequest {
		t.Errorf("err.Status() = %v, expected %v", err.Status(), http.StatusBadRequest)
	}
}

func TestRun(t *testing.T) {
}
