package httputil

import (
	"log"
	"net/http"
)

type AccessLogger struct {
	h http.Handler
}

func (a *AccessLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Handled", r.Method, r.URL.Path)
	a.h.ServeHTTP(w, r)
}

func NewAccessLogger(h http.Handler) http.Handler {
	return &AccessLogger{h}
}
