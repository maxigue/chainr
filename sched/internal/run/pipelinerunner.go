package run

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/pipeline"
)

type pipelineRunner struct {
	Pipeline pipeline.Pipeline
}

func (r *pipelineRunner) Run() {
	log.Println("PipelineRunner.Run()", r.Pipeline)
}

func newPipelineRunner(spec []byte) (runner, httputil.ErrorWithStatus) {
	var p pipeline.Pipeline
	err := json.Unmarshal(spec, &p)
	if err != nil {
		return nil, httputil.NewErrorWithStatus(err, http.StatusBadRequest)
	}
	return &pipelineRunner{p}, nil
}
