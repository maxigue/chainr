package httputil

import "net/http"

type ResponseBody struct {
	Kind  string                   `json:"kind"`
	Links map[string]*ResponseLink `json:"links"`
}

type ResponseLink struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

func NewResponseBody(r *http.Request, kind string) *ResponseBody {
	rb := ResponseBody{
		Kind:  kind,
		Links: make(map[string]*ResponseLink),
	}

	rb.Links["self"] = NewResponseLink(r, r.URL.RequestURI(), "Link to the current resource")
	return &rb
}

func NewResponseLink(r *http.Request, uri string, desc string) *ResponseLink {
	url := r.Host + uri
	return &ResponseLink{url, desc}
}
