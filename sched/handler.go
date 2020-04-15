package main

import (
	"encoding/json"
	"net/http"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/run"
)

type apiResource struct {
	Resource    string
	Description string
	Handler     http.Handler
}

type apiHandler struct {
	resources []apiResource
}

func NewHandler() http.Handler {
	mux := httputil.NewServeMux()
	apiResources := []apiResource{
		apiResource{"runs", "Interact with runs", run.NewHandler()},
	}
	for _, res := range apiResources {
		mux.Handle("/api/"+res.Resource, res.Handler)
		mux.Handle("/api/"+res.Resource+"/", res.Handler)
	}
	mux.Handle("/api", &apiHandler{apiResources})
	return httputil.NewAccessLogger(mux)
}

func (h *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		httputil.WriteError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.get(w, r)
}

func (h *apiHandler) get(w http.ResponseWriter, r *http.Request) {
	resp := httputil.NewResponseBody(r, "APIResourceList")
	for _, res := range h.resources {
		resp.Links[res.Resource] = httputil.NewResponseLink("/api/"+res.Resource, res.Description)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		httputil.WriteError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
