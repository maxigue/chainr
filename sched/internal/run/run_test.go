package run

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Tyrame/chainr/sched/internal/pipeline"
)

func TestRunHandler(t *testing.T) {
	Convey("Scenario: run a pipeline", t, func() {
		Convey("Given a pipeline is run", func() {
			w := httptest.NewRecorder()
			handler := http.Handler(NewHandler())
			uri := "/api/runs"

			Convey("When the data is a valid pipeline", func() {
				pipeline := pipeline.Pipeline{}
				b, err := json.Marshal(pipeline)
				if err != nil {
					t.Fatal(err)
				}

				r, err := http.NewRequest("POST", uri, bytes.NewReader(b))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should succeed with code 202", nil)

				Convey("The response should have the right Content-Type", func() {
					So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				})
			})

			Convey("When there is no data", func() {
				r, err := http.NewRequest("POST", uri, strings.NewReader(""))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 400", func() {
					So(w.Code, ShouldEqual, 400)
				})
			})

			Convey("When the data is has an invalid format", func() {
				Convey("The request should fail with code 400", nil)
			})

			Convey("When the data has an invalid dependency tree", func() {
				Convey("The request should fail with code 422", nil)
			})

			Convey("When the method is not POST", func() {
				r, err := http.NewRequest("GET", uri, strings.NewReader(""))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 405", func() {
					So(w.Code, ShouldEqual, 405)
				})

				Convey("The response should have the Allow header", func() {
					So(w.Header().Get("Allow"), ShouldEqual, "POST")
				})

				Convey("The response should have the right Content-Type", func() {
					So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				})
			})
		})
	})
}
