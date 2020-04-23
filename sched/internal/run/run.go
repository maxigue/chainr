// Package run contains the representation of a run,
// along with HTTP handlers.
package run

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/pipeline"
)

type Run struct {
	Kind     string      `json:"kind"`
	Metadata RunMetadata `json:"metadata"`
}

type RunMetadata struct {
	SelfLink string    `json:"selfLink"`
	UID      uuid.UUID `json:"uid"`
}

func New() *Run {
	return &Run{
		Kind: "Run",
		Metadata: RunMetadata{
			UID: uuid.New(),
		},
	}
}

type RunList struct {
	Kind     string          `json:"kind"`
	Metadata RunListMetadata `json:"metadata"`
	Items    []Run           `json:"items"`
}

type RunListMetadata struct {
	SelfLink string `json:"selfLink"`
}

func NewList() *RunList {
	return &RunList{
		Kind: "RunList",
		Metadata: RunListMetadata{
			SelfLink: "/api/runs",
		},
		Items: make([]Run, 0),
	}
}

type runHandler struct{}

func NewHandler() http.Handler {
	mux := httputil.NewServeMux()
	mux.Handle("/api/runs", &runHandler{})
	mux.Handle("/api/runs/", &runHandler{})
	return mux
}

func (h *runHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		runUID := r.URL.Path[len("/api/runs"):]
		if len(runUID) == 0 {
			h.list(w)
		} else {
			h.get(w, runUID[1:])
		}
	case "POST":
		h.post(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		httputil.WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *runHandler) list(w http.ResponseWriter) {
	runList := NewList()
	httputil.WriteResponse(w, runList, http.StatusOK)
}

func (h *runHandler) get(w http.ResponseWriter, runUID string) {
	httputil.WriteError(w, "Run "+runUID+" was not found", http.StatusNotFound)
}

// This variable is here to be overridden by unit tests,
// to stub pipeline run and simulate errors.
var newPipeline = func(spec []byte) (pipeline.Pipeline, error) {
	return pipeline.NewFromSpec(spec)
}

func (h *runHandler) post(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httputil.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	p, err := newPipeline(body)
	if err != nil {
		httputil.WriteError(w, err, http.StatusBadRequest)
		return
	}

	run := New()
	if err := p.Run(run.Metadata.UID.String()); err != nil {
		log.Println("Pipeline run failed:", err.Error())
		httputil.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	run.Metadata.SelfLink = "/api/runs/" + run.Metadata.UID.String()
	httputil.WriteResponse(w, run, http.StatusAccepted)
}
