// Package run contains the representation of a run,
// along with HTTP handlers.
package run

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/pipeline"
)

type runHandler struct {
}

func NewHandler() http.Handler {
	mux := httputil.NewServeMux()
	mux.Handle("/api/runs", &runHandler{})
	return mux
}

func (h *runHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, httputil.NewError(r, "Method not allowed").String(), http.StatusMethodNotAllowed)
		return
	}

	h.post(w, r)
}

func (h *runHandler) post(w http.ResponseWriter, r *http.Request) {
	var pipeline pipeline.Pipeline
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, httputil.NewError(r, err.Error()).String(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &pipeline)
	if err != nil {
		log.Println("Unable to decode request body:", err.Error())
		log.Println("Body:", string(body))
		http.Error(w, httputil.NewError(r, err.Error()).String(), http.StatusBadRequest)
		return
	}
	log.Println(pipeline)
}
