package httputil

import (
	"testing"

	"net/http"
)

func TestNewResponseBody(t *testing.T) {
	rb := NewResponseBody("Test")
	if rb.Kind != "Test" {
		t.Errorf("rb.Kind = %v, expected %v", rb.Kind, "Test")
	}
	if rb.Links == nil {
		t.Errorf("rb.Links should not be nil")
	}
}

func TestNewSelfLink(t *testing.T) {
	r, err := http.NewRequest("GET", "/test", nil)
	r.Host = "localhost"
	if err != nil {
		t.Fatal(err)
	}

	sl := NewSelfLink(r)
	if sl.URL != "localhost/test" {
		t.Errorf("sl.URL = %v, expected %v", sl.URL, "localhost/test")
	}
	dfltDesc := "Link to the current resource"
	if sl.Description != dfltDesc {
		t.Errorf("sl.Description = %v, expected %v", sl.Description, dfltDesc)
	}
}
