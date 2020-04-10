package pipeline

import (
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/Tyrame/chainr/sched/config"
)

func TestRunHandler(t *testing.T) {
	r, err := http.NewRequest("POST", "/pipeline/run", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler := http.Handler(NewRunHandler(config.Configuration{}))
	handler.ServeHTTP(w, r)
}
