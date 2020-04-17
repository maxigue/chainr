package httputil

import (
	"testing"

	"errors"
	"net/http"
)

func TestNewErrorWithStatus(t *testing.T) {
	e := NewErrorWithStatus(errors.New("Test"), http.StatusBadRequest)
	if e.Error() != "Test" {
		t.Errorf("e.Error() = %v, expected Test", e.Error())
	}
	if e.Status() != http.StatusBadRequest {
		t.Errorf("e.Status() = %v, expected %v", e.Status(), http.StatusBadRequest)
	}
}
