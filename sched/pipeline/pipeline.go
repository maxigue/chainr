// Package pipeline contains the code handling pipelines, including HTTP
// handlers and the whole pipeline representation and execution.
package pipeline

import (
	"net/http"

	"github.com/Tyrame/chainr/sched/config"
)

type Pipeline struct {
	Jobs map[string]Job
}

type Job struct {
	Image string
	Run string
	DependsOn []JobDependency
}

type JobDependency struct {
	Job string
	Conditions []ConditionDependency
}

type ConditionDependency struct {
	Failure bool
}

func NewHandler(cfg config.Configuration) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/pipeline/run", NewRunHandler(cfg))
	return mux
}
