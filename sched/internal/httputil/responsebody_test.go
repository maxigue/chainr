package httputil

import (
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestWriteResponse(t *testing.T) {
	w := httptest.NewRecorder()
	uri := "/test"
	r, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	body := NewResponseBody(r, "Test")
	WriteResponse(w, r, body)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("contentType = %v, expected application/json", contentType)
	}

	var resp ResponseBody
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Kind != "Test" {
		t.Errorf("resp.Kind = %v, expected Test", resp.Kind)
	}
	if resp.Metadata != nil {
		t.Errorf("resp.Metadata = %v, expected nil", resp.Metadata)
	}
}
