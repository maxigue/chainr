package httputil

import (
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func TestNewServeMux(t *testing.T) {
	mux := NewServeMux()
	r, err := http.NewRequest("GET", "/notexisting", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler, _ := mux.Handler(r)

	handler.ServeHTTP(w, r)
	var error Error
	json.NewDecoder(w.Body).Decode(&error)

	if w.Code != http.StatusNotFound {
		t.Errorf("w.Code = %v, expected %v", w.Code, http.StatusNotFound)
	}
	if error.Kind != "Error" {
		t.Errorf("error.Kind = %v, expected Error", error.Kind)
	}
}
