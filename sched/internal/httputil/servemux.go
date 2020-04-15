package httputil

import "net/http"

func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, NewError(r, "Resource not found").String(), http.StatusNotFound)
		}
	})
	return mux
}
