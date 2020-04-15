package httputil

import (
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func TestNewError(t *testing.T) {
	uri := "/test"
	r, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	marshaledErr := NewError(r, "testError").Bytes()

	var errJSON Error
	err = json.Unmarshal(marshaledErr, &errJSON)
	if err != nil {
		t.Fatal(err.Error())
	}

	if errJSON.Kind != "Error" {
		t.Errorf("errJSON.Kind = %v, expected %v", errJSON.Kind, "Error")
	}
	if errJSON.Error != "testError" {
		t.Errorf("errJSON.Error = %v, expected %v", errJSON.Error, "testError")
	}
	if errJSON.Links["self"].URL != uri {
		t.Errorf("errJSON.Links[self].URL = %v, expected %v", errJSON.Links["test"].URL, uri)
	}
}

func TestErrorString(t *testing.T) {
	r, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	e := NewError(r, "testError")
	str := e.String()
	if len(str) == 0 {
		t.Error("len(str) = 0, expected > 0")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	WriteError(w, r, "testError", http.StatusBadRequest)
	var resp ResponseBody
	json.NewDecoder(w.Body).Decode(&resp)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("contentType = %v, expected application/json", contentType)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("w.Code = %v, expected %v", w.Code, http.StatusBadRequest)
	}
	if resp.Kind != "Error" {
		t.Errorf("resp.Kind = %v, expected Error", resp.Kind)
	}
}
