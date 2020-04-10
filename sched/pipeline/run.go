package pipeline

import (
	"log"
	"net/http"

	"github.com/Tyrame/chainr/sched/config"
)

type RunHandler struct {
	cfg config.Configuration
}

func NewRunHandler(cfg config.Configuration) http.Handler {
	return &RunHandler{cfg}
}

func (h *RunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Handled (run)", r.Method, r.URL.Path)
}
