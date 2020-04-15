package httputil

import "net/http"

// The ServeMux adds a 404 handler on http's ServeMux.
func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			WriteError(w, r, "Resource not found", http.StatusNotFound)
		}
	})
	return mux
}
