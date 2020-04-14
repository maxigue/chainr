package pipeline

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"

	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Tyrame/chainr/sched/config"
)

func TestRunHandler(t *testing.T) {
	Convey("Scenario: run a pipeline", t, func() {
		Convey("Given a pipeline is run", func() {
			w := httptest.NewRecorder()
			handler := http.Handler(NewRunHandler(config.Configuration{}))

			Convey("When there is no data", func() {
				r, err := http.NewRequest("POST", "/pipeline/run", strings.NewReader(""))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("Then the request should fail with 400", func() {
					So(w.Code, ShouldEqual, 400)
				})
			})
		})
	})
}
