package main

import "net/http"

type Handler struct {
	cfg Configuration
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func NewHandler(cfg Configuration) http.Handler {
	return Handler{cfg}
}
