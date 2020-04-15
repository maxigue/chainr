package httputil

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	ResponseBody
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
