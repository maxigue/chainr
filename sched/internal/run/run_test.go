package run

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
)

func TestNew(t *testing.T) {
	spec := []byte(`{
		"kind": "Pipeline",
		"jobs": {
			"job1": {
				"image": "busybox",
				"run": "exit 0"
			}
		}
	}`)

	r, err := New(spec)
	if err != nil {
		t.Fatal(err)
	}
	if r.Kind != "Run" {
		t.Errorf("r.Kind = %v, expected Run", r.Kind)
	}
	selfLink := "/api/runs/" + r.Metadata.UID
	if r.Metadata.SelfLink != selfLink {
		t.Errorf("r.Metadata.SelfLink = %v, expected %v", r.Metadata.SelfLink, selfLink)
	}
}

func TestNewFail(t *testing.T) {
	spec := []byte(`{invalid}`)

	_, err := New(spec)
	if err == nil {
		t.Errorf("err is nil, expected non-nil")
	}
}

func TestNewList(t *testing.T) {
	items := map[string]Status{
		"run1": Status{
			"job1": "PENDING",
			"job2": "RUNNING",
		},
	}

	l := NewList(items)
	if l.Kind != "RunList" {
		t.Errorf("l.Kind = %v, expected RunList", l.Kind)
	}
	if l.Metadata.SelfLink != "/api/runs" {
		t.Errorf("l.Metadata.SelfLink = %v, expected /api/runs", l.Metadata.SelfLink)
	}
	if len(l.Items) != 1 {
		t.Errorf("len(l.Items) = %v, expected 1", len(l.Items))
	}
	if l.Items[0].Metadata.UID != "run1" {
		t.Errorf("l.Items[0].Metadata.UID = %v, expected run1", l.Items[0].Metadata.UID)
	}
	if l.Items[0].Metadata.SelfLink != "/api/runs/run1" {
		t.Errorf("l.Items[0].Metadata.SelfLink = %v, expected /api/runs/run1", l.Items[0].Metadata.SelfLink)
	}
	if l.Items[0].Status["job2"] != "RUNNING" {
		t.Errorf("l.Items[0].Status[job2] = %v, expected RUNNING", l.Items[0].Status["job2"])
	}
}

func TestNewHandler(t *testing.T) {
	h := NewHandler()
	if h == nil {
		t.Errorf("h is nil, expected non-nil")
	}
}

type nonEmptyScheduler struct{}

func (s nonEmptyScheduler) Schedule(run Run) (Status, error) {
	return Status{}, nil
}
func (s nonEmptyScheduler) Status(runUID string) (Status, error) {
	return Status{
		"job1": "RUNNING",
		"job2": "PENDING",
	}, nil
}
func (s nonEmptyScheduler) StatusMap() (map[string]Status, error) {
	return map[string]Status{
		"run1": Status{
			"job1": "RUNNING",
			"job2": "PENDING",
		},
	}, nil
}

type emptyScheduler struct{}

func (s emptyScheduler) Schedule(run Run) (Status, error) {
	return Status{}, nil
}
func (s emptyScheduler) Status(runUID string) (Status, error) {
	return Status{}, nil
}
func (s emptyScheduler) StatusMap() (map[string]Status, error) {
	return make(map[string]Status), nil
}

type failingScheduler struct{}

func (s failingScheduler) Schedule(run Run) (Status, error) {
	return Status{}, errors.New("fail")
}
func (s failingScheduler) Status(runUID string) (Status, error) {
	return Status{}, errors.New("fail")
}
func (s failingScheduler) StatusMap() (map[string]Status, error) {
	return make(map[string]Status), errors.New("fail")
}

type notFoundScheduler struct {
	failingScheduler
}

func (s notFoundScheduler) Status(runUID string) (Status, error) {
	return Status{}, &NotFoundError{runUID}
}

func TestRunHandlerList(t *testing.T) {
	Convey("Scenario: list runs", t, func() {
		Convey("Given the runs list is requested", func() {
			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "/api/runs", nil)
			if err != nil {
				t.Fatal(err)
			}

			Convey("When there are runs", func() {
				handler := http.Handler(newHandler(&nonEmptyScheduler{}))
				handler.ServeHTTP(w, r)
				var runList RunList
				json.NewDecoder(w.Body).Decode(&runList)

				Convey("The response should succeed with code 200", func() {
					So(w.Code, ShouldEqual, 200)
				})

				Convey("The response should be of kind RunList", func() {
					So(runList.Kind, ShouldEqual, "RunList")
				})

				Convey("The response should have items", func() {
					So(runList.Items, ShouldNotBeEmpty)
				})

				Convey("Response items should have a status for each job", func() {
					So(runList.Items[0].Status["job1"], ShouldEqual, "RUNNING")
					So(runList.Items[0].Status["job2"], ShouldEqual, "PENDING")
				})
			})

			Convey("When there are no runs", func() {
				handler := http.Handler(newHandler(&emptyScheduler{}))
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

			Convey("When the scheduler fails", func() {
				handler := http.Handler(newHandler(&failingScheduler{}))
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 500", func() {
					So(w.Code, ShouldEqual, 500)
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

			Convey("When the run exists", func() {
				handler := http.Handler(newHandler(&nonEmptyScheduler{}))
				handler.ServeHTTP(w, r)
				var run Run
				json.NewDecoder(w.Body).Decode(&run)

				Convey("The request should succeed with code 200", func() {
					So(w.Code, ShouldEqual, 200)
				})

				Convey("The response should have a status for each job", func() {
					So(run.Status["job1"], ShouldEqual, "RUNNING")
					So(run.Status["job2"], ShouldEqual, "PENDING")
				})
			})

			Convey("When the run does not exist", func() {
				handler := http.Handler(newHandler(&notFoundScheduler{}))
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 404", func() {
					So(w.Code, ShouldEqual, 404)
				})
			})

			Convey("When the scheduler fails", func() {
				handler := http.Handler(newHandler(&failingScheduler{}))
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 500", func() {
					So(w.Code, ShouldEqual, 500)
				})
			})
		})
	})
}

type errorReader struct{}

func (r errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("fail")
}

func TestRunHandlerPost(t *testing.T) {
	Convey("Scenario: run a pipeline", t, func() {
		Convey("Given a run is created", func() {
			w := httptest.NewRecorder()
			handler := http.Handler(newHandler(&emptyScheduler{}))
			uri := "/api/runs"

			Convey("When the data is a valid pipeline", func() {
				body := `{
					"kind": "Pipeline",
					"jobs": {}
				}`
				r, err := http.NewRequest("POST", uri, strings.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}

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

				Convey("And the run scheduling fails", func() {
					handler = http.Handler(newHandler(&failingScheduler{}))
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

			Convey("When the body cannot be read", func() {
				r, err := http.NewRequest("POST", uri, &errorReader{})
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 500", func() {
					So(w.Code, ShouldEqual, 500)
				})
			})
		})
	})
}
