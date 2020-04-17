package httputil

import (
	"encoding/json"
	"net/http"
)

// This structure represents a body that can be returned as a response.
// TODO: This design is poor, and the concept of Kind must be improved.
type ResponseBody struct {
	Kindable
	Metadata interface{}              `json:"metadata,omitempty"`
	Links    map[string]*ResponseLink `json:"links"`
}

type ResponseLink struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

func NewResponseBody(r *http.Request, kind string) *ResponseBody {
	rb := ResponseBody{}
	rb.Kindable = Kindable{kind}
	rb.Links = make(map[string]*ResponseLink)

	rb.Links["self"] = NewResponseLink(r.URL.RequestURI(), "Link to the current resource")
	return &rb
}

func NewResponseLink(uri string, desc string) *ResponseLink {
	return &ResponseLink{uri, desc}
}

func WriteResponse(w http.ResponseWriter, r *http.Request, body *ResponseBody) {
	bytes, err := json.Marshal(body)
	if err != nil {
		WriteError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
