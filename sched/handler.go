package main

import (
	"net/http"

	"github.com/Tyrame/chainr/sched/config"
	"github.com/Tyrame/chainr/sched/pipeline"
)

func NewHandler(cfg config.Configuration) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/pipeline/", pipeline.NewHandler(cfg))
	return NewAccessLogger(mux)
}
