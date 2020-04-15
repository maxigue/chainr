package httputil

import (
	"encoding/json"
	"net/http"
)

// Error represents the body returned in case of error.
type Error struct {
	*ResponseBody
	Error string `json:"error"`
}

func NewError(r *http.Request, err string) *Error {
	e := new(Error)
	e.ResponseBody = NewResponseBody(r, "Error")
	e.Error = err
	return e
}

func (e *Error) Bytes() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		return []byte("{}")
	}
	return bytes
}

func (e *Error) String() string {
	return string(e.Bytes())
}

// Writes an error json containing the error string, and the links.
// Also sets the response headers.
func WriteError(w http.ResponseWriter, r *http.Request, err string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	e := NewError(r, err)
	w.Write(e.Bytes())
}
