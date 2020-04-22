package httputil

import (
	"testing"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func TestWriteResponse(t *testing.T) {
	w := httptest.NewRecorder()
	type body struct {
		A string `json:"a"`
	}
	b := body{"foo"}

	WriteResponse(w, b, http.StatusOK)

	if w.Code != http.StatusOK {
		t.Errorf("w.Code = %v, expected %v", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("contentType = %v, expected application/json", contentType)
	}

	bytes, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}

	s := string(bytes)
	expected := `{"a":"foo"}`
	if s != expected {
		t.Errorf("s = %v, expected %v", s, expected)
	}
}

func TestWriteResponseError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteResponse(w, func() {}, http.StatusOK)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("w.Code = %v, expected %v", w.Code, http.StatusInternalServerError)
	}

	var error Error
	json.NewDecoder(w.Body).Decode(&error)
	if error.Kind != "Error" {
		t.Errorf("error.Kind = %v, expected Error", error.Kind)
	}
}
