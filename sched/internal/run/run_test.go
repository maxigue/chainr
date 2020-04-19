package run

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/k8s"
	"github.com/Tyrame/chainr/sched/internal/pipeline"
)

func TestRunHandler(t *testing.T) {
	Convey("Scenario: run a pipeline", t, func() {
		Convey("Given a pipeline is run", func() {
			w := httptest.NewRecorder()
			handler := http.Handler(NewHandler(k8s.NewStub()))
			uri := "/api/runs"

			Convey("When the data is a valid pipeline", func() {
				pipeline := pipeline.New()
				b, err := json.Marshal(pipeline)
				if err != nil {
					t.Fatal(err)
				}

				r, err := http.NewRequest("POST", uri, bytes.NewReader(b))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)
				var resp httputil.ResponseBody
				json.NewDecoder(w.Body).Decode(&resp)

				Convey("The request should succeed with code 202", func() {
					So(w.Code, ShouldEqual, 202)
				})

				Convey("The response should have the Content-Type application/json", func() {
					So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				})

				Convey("The response should be of kind Run", func() {
					So(resp.Kind, ShouldEqual, "Run")
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

			Convey("When the data has an invalid format", func() {
				body := `{wrongformat}`
				r, err := http.NewRequest("POST", uri, strings.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 400", func() {
					So(w.Code, ShouldEqual, 400)
				})
			})

			Convey("When the data is a Pipeline, but does not meet the Pipeline schema", func() {
				body := `{"kind": "Pipeline"}`
				r, err := http.NewRequest("POST", uri, strings.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 400", func() {
					So(w.Code, ShouldEqual, 400)
				})
			})

			Convey("When the data has an unsupported kind", func() {
				body := `{"kind": "Notexisting"}`
				r, err := http.NewRequest("POST", uri, strings.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 400", func() {
					So(w.Code, ShouldEqual, 400)
				})
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

				Convey("The response should have the Content-Type application/json", func() {
					So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				})
			})
		})
	})
}
