package httputil

import (
	"testing"

	"net/http"
)

func TestNewResponseBody(t *testing.T) {
	uri := "/test"
	r, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	rb := NewResponseBody(r, "Test")
	if rb.Kind != "Test" {
		t.Errorf("rb.Kind = %v, expected %v", rb.Kind, "Test")
	}
	if rb.Links == nil {
		t.Errorf("rb.Links should not be nil")
	}

	selfLink := rb.Links["self"]
	if selfLink.URL != uri {
		t.Errorf("selfLink.URL = %v, expected %v", selfLink.URL, uri)
	}
	dfltDesc := "Link to the current resource"
	if selfLink.Description != dfltDesc {
		t.Errorf("selfLink.Description = %v, expected %v", selfLink.Description, dfltDesc)
	}
}
