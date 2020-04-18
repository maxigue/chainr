package run

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Tyrame/chainr/sched/internal/httputil"
)

type runner interface {
	Run()
}

type runnerFactory func([]byte) (runner, httputil.ErrorWithStatus)

var runnerFactories = map[string]runnerFactory{
	"Pipeline": newPipelineRunner,
}

func newRunner(body []byte) (runner runner, error httputil.ErrorWithStatus) {
	var kindable httputil.Kindable
	err := json.Unmarshal(body, &kindable)
	if err != nil {
		log.Println("Unable to decode request body:", err.Error())
		log.Println("Body:", string(body))
		return nil, httputil.NewErrorWithStatus(err, http.StatusBadRequest)
	}

	kinds := make([]string, 0, len(runnerFactories))
	for kind, factory := range runnerFactories {
		kinds = append(kinds, kind)
		if kindable.Kind == kind {
			runner, errws := factory(body)
			if errws != nil {
				return nil, errws
			}
			return runner, nil
		}
	}

	err = errors.New("invalid kind " + kindable.Kind + ", expected " + strings.Join(kinds, ", "))
	return nil, httputil.NewErrorWithStatus(err, http.StatusBadRequest)
}
