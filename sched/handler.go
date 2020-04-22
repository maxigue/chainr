package main

import (
	"net/http"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/run"
)

type apiResourceList struct {
	Kind      string                 `json:"kind"`
	Metadata  metadata               `json:"metadata"`
	Resources map[string]apiResource `json:"resources"`
}

type metadata struct {
	SelfLink string `json:"selfLink"`
}

type apiResource struct {
	URL         string       `json:"url"`
	Description string       `json:"description"`
	Handler     http.Handler `json:"-"`
}

type apiHandler struct {
	lst *apiResourceList
}

func newApiResourceList() *apiResourceList {
	return &apiResourceList{
		Kind: "APIResourceList",
		Metadata: metadata{
			SelfLink: "/api",
		},
		Resources: map[string]apiResource{
			"runs": apiResource{"/api/runs", "Interact with runs", run.NewHandler()},
		},
	}
}

func NewHandler() http.Handler {
	mux := httputil.NewServeMux()
	lst := newApiResourceList()
	for _, res := range lst.Resources {
		mux.Handle(res.URL, res.Handler)
		mux.Handle(res.URL+"/", res.Handler)
	}
	mux.Handle("/api", &apiHandler{lst})
	return httputil.NewAccessLogger(mux)
}

func (h *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		httputil.WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.get(w, r)
}

func (h *apiHandler) get(w http.ResponseWriter, r *http.Request) {
	httputil.WriteResponse(w, h.lst, http.StatusOK)
}
