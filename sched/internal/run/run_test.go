package run

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Tyrame/chainr/sched/internal/httputil"
	"github.com/Tyrame/chainr/sched/internal/pipeline"
)

type validPipeline struct{}

func (p *validPipeline) Run(string) error {
	return nil
}
func newValidPipeline() pipeline.Pipeline {
	return &validPipeline{}
}

type invalidPipeline struct{}

func (p *invalidPipeline) Run(string) error {
	return errors.New("fail")
}
func newInvalidPipeline() pipeline.Pipeline {
	return &invalidPipeline{}
}

func TestRunHandler(t *testing.T) {
	Convey("Scenario: run a pipeline", t, func() {
		Convey("Given a pipeline is run", func() {
			w := httptest.NewRecorder()
			handler := http.Handler(NewHandler())
			uri := "/api/runs"

			Convey("When the data is a valid pipeline", func() {
				r, err := http.NewRequest("POST", uri, strings.NewReader(""))
				if err != nil {
					t.Fatal(err)
				}
				old := newPipeline
				newPipeline = func(spec []byte) (pipeline.Pipeline, error) {
					return newValidPipeline(), nil
				}
				defer func() {
					newPipeline = old
				}()

				Convey("The request should succeed with code 202", func() {
					handler.ServeHTTP(w, r)
					So(w.Code, ShouldEqual, 202)
				})

				Convey("The response should have the Content-Type application/json", func() {
					handler.ServeHTTP(w, r)
					So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				})

				Convey("The response should be of kind Run", func() {
					handler.ServeHTTP(w, r)
					var resp httputil.ResponseBody
					json.NewDecoder(w.Body).Decode(&resp)
					So(resp.Kind, ShouldEqual, "Run")
				})

				Convey("And the pipeline run fails", func() {
					newPipeline = func(spec []byte) (pipeline.Pipeline, error) {
						return newInvalidPipeline(), nil
					}
					handler.ServeHTTP(w, r)
					Convey("The request should fail with code 500", func() {
						So(w.Code, ShouldEqual, 500)
					})
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

func TestNew(t *testing.T) {
	r := New()
	if r.Kind != "Run" {
		t.Errorf("r.Kind = %v, expected Run", r.Kind)
	}
}
