package run

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

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

func TestRunHandlerList(t *testing.T) {
	Convey("Scenario: list runs", t, func() {
		Convey("Given the runs list is requested", func() {
			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "/api/runs", nil)
			if err != nil {
				t.Fatal(err)
			}
			handler := http.Handler(NewHandler())

			Convey("When there are runs", func() {
				handler.ServeHTTP(w, r)
				var runList RunList
				json.NewDecoder(w.Body).Decode(&runList)

				Convey("The response should succeed with code 200", func() {
					So(w.Code, ShouldEqual, 200)
				})

				Convey("The response should be of kind RunList", func() {
					So(runList.Kind, ShouldEqual, "RunList")
				})

				Convey("The response should have items", nil)

				Convey("Response items should have a global status", nil)

				Convey("Response items should have a status for each job", nil)
			})
			Convey("When there are no runs", func() {
				handler.ServeHTTP(w, r)
				var runList RunList
				json.NewDecoder(w.Body).Decode(&runList)

				Convey("The response should succeed with code 200", func() {
					So(w.Code, ShouldEqual, 200)
				})

				Convey("The response should be of kind RunList", func() {
					So(runList.Kind, ShouldEqual, "RunList")
				})

				Convey("The response should have empty items", func() {
					So(runList.Items, ShouldBeEmpty)
				})
			})
		})
	})
}

func TestRunHandlerGet(t *testing.T) {
	Convey("Scenario: get a single run", t, func() {
		Convey("Given a run is requested", func() {
			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "/api/runs/abc", nil)
			if err != nil {
				t.Fatal(err)
			}
			handler := http.Handler(NewHandler())

			Convey("When the run exists", func() {
				handler.ServeHTTP(w, r)

				Convey("The request should succeed with code 200", nil)

				Convey("The response should have a global status", nil)

				Convey("The response should have a status for each job", nil)
			})

			Convey("When the run does not exist", func() {
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 404", func() {
					So(w.Code, ShouldEqual, 404)
				})
			})
		})
	})
}

func TestRunHandlerPost(t *testing.T) {
	Convey("Scenario: run a pipeline", t, func() {
		Convey("Given a run is created", func() {
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
					var run Run
					json.NewDecoder(w.Body).Decode(&run)
					So(run.Kind, ShouldEqual, "Run")
				})

				Convey("The response should have a link to the created run", func() {
					handler.ServeHTTP(w, r)
					var run Run
					json.NewDecoder(w.Body).Decode(&run)
					So(run.Metadata.SelfLink, ShouldStartWith, "/api/runs/")
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

			Convey("When the method is unsupported", func() {
				r, err := http.NewRequest("DELETE", uri, strings.NewReader(""))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 405", func() {
					So(w.Code, ShouldEqual, 405)
				})

				Convey("The response should have the Allow header", func() {
					So(w.Header().Get("Allow"), ShouldEqual, "GET, POST")
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
