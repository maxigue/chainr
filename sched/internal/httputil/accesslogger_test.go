package httputil

import (
	"testing"

	"net/http"
	"net/http/httptest"
)

type DummyHandler struct{}

func (h DummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Test-Ok", "Ok")
}

func TestAccessLogger(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler := http.Handler(NewAccessLogger(DummyHandler{}))
	handler.ServeHTTP(w, r)

	if len(w.Header()) != 1 {
		t.Fatalf("len(w.Header()) = %v, expected 1", len(w.Header()))
	}
	if w.Header().Get("Test-Ok") != "Ok" {
		t.Errorf("w.Header().Get(Test-Ok) = %v, expected Ok", w.Header().Get("Test-Ok"))
	}
}
