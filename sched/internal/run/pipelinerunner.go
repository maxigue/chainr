package run

import (
	"log"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/pipeline"
)

type pipelineRunner struct {
	Pipeline *pipeline.Pipeline
}

func newPipelineRunner(spec []byte) (runner, httputil.ErrorWithStatus) {
	p, errws := pipeline.NewFromSpec(spec)
	if errws != nil {
		return nil, errws
	}
	return &pipelineRunner{p}, nil
}

func (r *pipelineRunner) Run() {
	log.Println("PipelineRunner.Run()", r.Pipeline)
}
