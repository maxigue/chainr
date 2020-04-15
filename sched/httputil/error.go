package httputil

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	ResponseBody
	Error string `json:"error"`
}

func NewError(r *http.Request, e string) string {
	var errStruct Error
	errStruct.ResponseBody = NewResponseBody(r, "Error")
	errStruct.Error = e
	bytes, err := json.Marshal(errStruct)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
