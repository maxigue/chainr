// Package run contains the representation of a run,
// along with HTTP handlers.
package run

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
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

	var kindable httputil.Kindable
	err = json.Unmarshal(body, &kindable)
	if err != nil {
		log.Println("Unable to decode request body:", err.Error())
		log.Println("Body:", string(body))
		httputil.WriteError(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	switch kindable.Kind {
	case "Pipeline":
		err = h.runPipeline(body)
	default:
		err = errors.New("invalid kind " + kindable.Kind + ", expected Pipeline")
	}

	if err != nil {
		httputil.WriteError(w, r, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run := NewRun()
	resp := httputil.NewResponseBody(r, "Run")
	resp.Metadata = run.Metadata
	// TODO: links are not implemented yet
	delete(resp.Links, "self")
	// resp.Links["self"].URL = "/api/runs/" + run.UID.String()
	// resp.Links["status"].URL = "/api/runs/" + run.UID.String() + "/status"
	w.WriteHeader(http.StatusAccepted)
	httputil.WriteResponse(w, r, resp)
}

func (h *runHandler) runPipeline(body []byte) error {
	return nil
}
