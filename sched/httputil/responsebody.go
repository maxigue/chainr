package httputil

import "net/http"

type ResponseBody struct {
	Kind string                   `json:"kind"`
	Links map[string]ResponseLink `json:"links"`
}

type ResponseLink struct {
	URL string         `json:"url"`
	Description string `json:"description"`
}

func NewResponseBody(kind string) ResponseBody {
	return ResponseBody{
		Kind: kind,
		Links: make(map[string]ResponseLink),
	}
}

func NewSelfLink(r *http.Request) ResponseLink {
	url := r.Host + r.URL.RequestURI()
	return ResponseLink{url, "Link to the current resource"}
}
