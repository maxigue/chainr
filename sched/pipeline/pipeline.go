// Package pipeline contains the code handling pipelines, including HTTP
// handlers and the whole pipeline representation and execution.
package pipeline

import (
	"net/http"

	"github.com/Tyrame/chainr/sched/config"
)

func NewHandler(cfg config.Configuration) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/pipeline/run", NewRunHandler(cfg))
	return mux
}
