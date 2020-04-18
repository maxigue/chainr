// Package run contains the representation of a run,
// along with HTTP handlers.
package run

import (
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"

	"github.com/Tyrame/chainr/sched/internal/httputil"
)

type Run struct {
	httputil.Kindable
	Metadata RunMetadata `json:"metadata"`
}

type RunMetadata struct {
	UID uuid.UUID `json:"uid"`
}

func NewRun() *Run {
	return &Run{
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
		httputil.WriteError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.post(w, r)
}

func (h *runHandler) post(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httputil.WriteError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: this is over-engineered, because only pipelines can be run.
	// It would be better to improve pipeline.newFromSpec() to manage PipelineLinks.
	// It would make the PipelineLink kind a part of the pipeline package.
	// It would also remove the need of an error with status.
	// This would give:
	//     p, err := pipeline.NewFromSpec(body)
	//     if err != nil {
	//         httputil.WriteError(w, r, err.Error(), http.StatusBadRequest)
	//         return
	//     }
	//     if err := p.Validate(); err != nil {
	//         httputil.WriteError(w, r, err.Error(), http.StatusUnprocessableEntity)
	//         return
	//     }
	//     runner := newRunner(p)
	//
	// Note that with this approach, there is no longer a need for a runner, because
	// the run part can be part of the Pipeline itself.
	// Instead of:
	//     runner := newRunner(p)
	//     go runner.Run()
	// We can do:
	//     go p.Run()
	runner, errws := newRunner(body)
	if errws != nil {
		httputil.WriteError(w, r, errws.Error(), errws.Status())
		return
	}

	go runner.Run()

	run := NewRun()
	resp := httputil.NewResponseBody(r, "Run")
	resp.Metadata = run.Metadata
	delete(resp.Links, "self") // TODO: links are not implemented yet
	// resp.Links["self"].URL = "/api/runs/" + run.UID.String()
	// resp.Links["status"].URL = "/api/runs/" + run.UID.String() + "/status"
	w.WriteHeader(http.StatusAccepted)
	httputil.WriteResponse(w, r, resp)
}
