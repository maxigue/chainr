package httputil

import "net/http"

// Error represents the body returned in case of error.
type Error struct {
	Kind  string `json:"kind"`
	Error string `json:"error"`
}

func NewError(err string) Error {
	return Error{
		Kind:  "Error",
		Error: err,
	}
}

// Writes an error json containing the error string.
// It supports errors in format error and string.
func WriteError(w http.ResponseWriter, err interface{}, code int) {
	var str string
	switch e := err.(type) {
	case string:
		str = e
	case error:
		str = e.Error()
	}

	WriteResponse(w, NewError(str), code)
}
