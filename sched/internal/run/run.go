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
	UID uuid.UUID `json:"uid"`
}

func New() *Run {
	return &Run{
		Kind: "Run",
		Metadata: RunMetadata{
			UID: uuid.New(),
		},
	}
}

type runHandler struct {
}

func NewHandler() http.Handler {
	mux := httputil.NewServeMux()
	mux.Handle("/api/runs", &runHandler{})
	return mux
}

func (h *runHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		httputil.WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.post(w, r)
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

	// resp.Metadata.selfLink = "/api/runs/" + run.UID.String()
	httputil.WriteResponse(w, run, http.StatusAccepted)
}
