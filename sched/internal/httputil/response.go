package httputil

import (
	"encoding/json"
	"net/http"
)

// Writes a response with status code as application/json.
// If the body can not be marshalled to json, returns an internal server error.
func WriteResponse(w http.ResponseWriter, body interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")

	bytes, err := json.Marshal(body)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(bytes)
}
