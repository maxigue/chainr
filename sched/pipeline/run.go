package pipeline

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Tyrame/chainr/sched/config"
	"github.com/Tyrame/chainr/sched/httputil"
)

type RunHandler struct {
	cfg config.Configuration
}

func NewRunHandler(cfg config.Configuration) http.Handler {
	return &RunHandler{cfg}
}

func (h *RunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, httputil.NewError(r, "Method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	h.post(w, r)
}

func (h *RunHandler) post(w http.ResponseWriter, r *http.Request) {
	var pipeline Pipeline
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, httputil.NewError(r, err.Error()), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &pipeline)
	if err != nil {
		log.Println("Unable to decode request body:", err.Error())
		log.Println("Body:", string(body))
		http.Error(w, httputil.NewError(r, err.Error()), http.StatusBadRequest)
		return
	}
	log.Println(pipeline)
}
