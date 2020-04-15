// Package pipeline contains the code handling pipelines, including HTTP
// handlers and the whole pipeline representation and execution.
package pipeline

import (
	"encoding/json"
	"net/http"

	"github.com/Tyrame/chainr/sched/config"
	"github.com/Tyrame/chainr/sched/httputil"
)

type Pipeline struct {
	Jobs map[string]Job
}

type Job struct {
	Image     string
	Run       string
	DependsOn []JobDependency
}

type JobDependency struct {
	Job        string
	Conditions []ConditionDependency
}

type ConditionDependency struct {
	Failure bool
}

type pipelineHandler struct {
	cfg config.Configuration
}

func NewHandler(cfg config.Configuration) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/pipeline", &pipelineHandler{cfg})
	mux.Handle("/api/pipeline/run", NewRunHandler(cfg))
	return mux
}

func (h *pipelineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, httputil.NewError(r, "Method not allowed").String(), http.StatusMethodNotAllowed)
		return
	}

	h.get(w, r)
}

func (h *pipelineHandler) get(w http.ResponseWriter, r *http.Request) {
	resp := httputil.NewResponseBody(r, "APIResource")
	resp.Links["run"] = httputil.NewResponseLink(r, "/api/pipeline/run", "Run a pipeline by sending a configuration")

	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, httputil.NewError(r, err.Error()).String(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
