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

// Possible values for run status:
// - PENDING
// - RUNNING
// - SUCCESSFUL
// - FAILED
// - CANCELED
// Possible values for jobs status:
// - PENDING
// - RUNNING
// - SUCCESSFUL
// - FAILED
type Run struct {
	p Pipeline

	Kind     string   `json:"kind"`
	Metadata Metadata `json:"metadata"`
	Status   string   `json:"status"`
	Jobs     []RunJob `json:"jobs"`
}

type Metadata struct {
	SelfLink string `json:"selfLink"`
	UID      string `json:"uid"`
}

type RunJob struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Creates a run from a pipeline.
func New(p Pipeline) Run {
	uid := uuid.New().String()

	return Run{
		p: p,

		Kind: "Run",
		Metadata: Metadata{
			SelfLink: "/api/runs/" + uid,
			UID:      uid,
		},
	}
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
	Status   string   `json:"status"`
	Jobs     []RunJob `json:"jobs"`
}

func NewList(items []StatusListItem) RunList {
	list := RunList{
		Kind: "RunList",
		Metadata: RunListMetadata{
			SelfLink: "/api/runs",
		},
		Items: make([]RunListItem, 0, len(items)),
	}

	for _, item := range items {
		list.Items = append(list.Items, newListItem(item))
	}

	return list
}

func newListItem(item StatusListItem) RunListItem {
	return RunListItem{
		Metadata: Metadata{
			SelfLink: "/api/runs/" + item.RunUID,
			UID:      item.RunUID,
		},
		Status: item.Status.Run,
		Jobs:   item.Status.Jobs,
	}
}

type runHandler struct {
	pf    PipelineFactory
	sched Scheduler
}

func NewHandler() http.Handler {
	return newHandler(
		NewPipelineFactory(),
		NewScheduler(),
	)
}
func newHandler(pf PipelineFactory, sched Scheduler) http.Handler {
	handler := &runHandler{pf, sched}

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
	statusList, err := h.sched.StatusList()
	if err != nil {
		log.Println("Unable to get status map:", err.Error())
		httputil.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	runList := NewList(statusList)
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
		Status: status.Run,
		Jobs:   status.Jobs,
	}

	httputil.WriteResponse(w, run, http.StatusOK)
}

func (h *runHandler) post(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httputil.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	p, err := h.pf.Create(body)
	if err != nil {
		httputil.WriteError(w, err, http.StatusBadRequest)
		return
	}
	run := New(p)

	status, err := h.sched.Schedule(run)
	if err != nil {
		log.Println("Run scheduling failed:", err.Error())
		httputil.WriteError(w, err, http.StatusInternalServerError)
		return
	}

	run.Status = status.Run
	run.Jobs = status.Jobs

	httputil.WriteResponse(w, run, http.StatusAccepted)
}
