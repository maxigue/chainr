package main

import (
	"encoding/json"
	"net/http"

	"github.com/Tyrame/chainr/sched/config"
	"github.com/Tyrame/chainr/sched/httputil"
	"github.com/Tyrame/chainr/sched/pipeline"
)

type apiResource struct {
	Resource    string
	Description string
	Handler     http.Handler
}

type apiHandler struct {
	resources []apiResource
}

func NewHandler(cfg config.Configuration) http.Handler {
	mux := http.NewServeMux()
	apiResources := []apiResource{
		apiResource{"pipeline", "Interact with pipelines", pipeline.NewHandler(cfg)},
	}
	for _, res := range apiResources {
		mux.Handle("/api/"+res.Resource, res.Handler)
		mux.Handle("/api/"+res.Resource+"/", res.Handler)
	}
	mux.Handle("/api", &apiHandler{apiResources})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, httputil.NewError(r, "Resource not found"), http.StatusNotFound)
	})
	return httputil.NewAccessLogger(mux)
}

func (h *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, httputil.NewError(r, "Method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	h.get(w, r)
}

func (h *apiHandler) get(w http.ResponseWriter, r *http.Request) {
	resp := httputil.NewResponseBody(r, "APIResources")
	for _, res := range h.resources {
		resp.Links[res.Resource] = httputil.NewResponseLink(r, "/api/"+res.Resource, res.Description)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, httputil.NewError(r, err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
