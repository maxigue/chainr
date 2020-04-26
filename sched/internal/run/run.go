// Package run contains the representations of a run,
// along with HTTP handlers.
// It also contains a scheduler to allow to notify
// workers that a run must be processed.
package run

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/Tyrame/chainr/sched/internal/httputil"
)

type Run struct {
	p Pipeline

	Kind     string   `json:"kind"`
	Metadata Metadata `json:"metadata"`
	Status   Status   `json:"status"`
}

type Metadata struct {
	SelfLink string `json:"selfLink"`
	UID      string `json:"uid"`
}

// Status key is the job name, value is the status.
type Status map[string]string

// Creates a run from a JSON spec given as an array of bytes.
// If the spec has an invalid format, an error is returned.
func New(spec []byte) (Run, error) {
	p, err := NewPipeline(spec)
	if err != nil {
		return Run{}, err
	}

	uid := uuid.New().String()

	return Run{
		p: p,

		Kind: "Run",
		Metadata: Metadata{
			SelfLink: "/api/runs/" + uid,
			UID:      uid,
		},
	}, nil
}

type RunList struct {
	Kind     string          `json:"kind"`
	Metadata RunListMetadata `json:"metadata"`
	Items    []RunListItem   `json:"items"`
}

type RunListMetadata struct {
	SelfLink string `json:"selfLink"`
}

type RunListItem struct {
	Metadata Metadata `json:"metadata"`
	Status   Status   `json:"status"`
}

func NewList(items map[string]Status) RunList {
	list := RunList{
		Kind: "RunList",
		Metadata: RunListMetadata{
			SelfLink: "/api/runs",
		},
		Items: make([]RunListItem, 0, len(items)),
	}

	for runUID, status := range items {
		list.Items = append(list.Items, NewListItem(runUID, status))
	}

	return list
}

func NewListItem(runUID string, status Status) RunListItem {
	return RunListItem{
		Metadata: Metadata{
			SelfLink: "/api/runs/" + runUID,
			UID:      runUID,
		},
		Status: status,
	}
}

type runHandler struct {
	sched Scheduler
}

func NewHandler() http.Handler {
	return newHandler(NewScheduler())
}
func newHandler(sched Scheduler) http.Handler {
	handler := &runHandler{sched}

	mux := httputil.NewServeMux()
	mux.Handle("/api/runs", handler)
	mux.Handle("/api/runs/", handler)
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
	statusMap, err := h.sched.StatusMap()
	if err != nil {
		log.Println("Unable to get status map:", err.Error())
		httputil.WriteError(w, err, http.StatusInternalServerError)
	}

	runList := NewList(statusMap)
	httputil.WriteResponse(w, runList, http.StatusOK)
}

func (h *runHandler) get(w http.ResponseWriter, runUID string) {
	status, err := h.sched.Status(runUID)
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			httputil.WriteError(w, err, http.StatusNotFound)
		default:
			httputil.WriteError(w, err, http.StatusInternalServerError)
		}
		return
	}

	run := Run{
		Kind: "Run",
		Metadata: Metadata{
			SelfLink: "/api/runs/" + runUID,
			UID:      runUID,
		},
		Status: status,
	}

	httputil.WriteResponse(w, run, http.StatusOK)
}

func (h *runHandler) post(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httputil.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	run, err := New(body)
	if err != nil {
		httputil.WriteError(w, err, http.StatusBadRequest)
		return
	}

	run.Status, err = h.sched.Schedule(run)
	if err != nil {
		log.Println("Run scheduling failed:", err.Error())
		httputil.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	httputil.WriteResponse(w, run, http.StatusAccepted)
}
