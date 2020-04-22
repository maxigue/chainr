package httputil

import (
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func TestNewError(t *testing.T) {
	error := NewError("testError")

	if error.Kind != "Error" {
		t.Errorf("error.Kind = %v, expected %v", error.Kind, "Error")
	}
	if error.Error != "testError" {
		t.Errorf("error.Error = %v, expected %v", error.Error, "testError")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, "testError", http.StatusBadRequest)
	var error Error
	json.NewDecoder(w.Body).Decode(&error)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("contentType = %v, expected application/json", contentType)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("w.Code = %v, expected %v", w.Code, http.StatusBadRequest)
	}
	if error.Kind != "Error" {
		t.Errorf("error.Kind = %v, expected Error", error.Kind)
	}
}
