package pipeline

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Tyrame/chainr/sched/config"
	"github.com/Tyrame/chainr/sched/httputil"
)

func TestHandler(t *testing.T) {
	Convey("Scenario: get pipeline API resource", t, func() {
		Convey("Given pipelines API resource is requested", func() {
			w := httptest.NewRecorder()
			handler := http.Handler(NewHandler(config.Configuration{}))
			uri := "/api/pipeline"

			Convey("When the method is GET", func() {
				r, err := http.NewRequest("GET", uri, strings.NewReader(""))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)
				var resp httputil.ResponseBody
				json.NewDecoder(w.Body).Decode(&resp)

				Convey("The request should succeed with 200", func() {
					So(w.Code, ShouldEqual, 200)
				})

				Convey("The response should contain an APIResource", func() {
					So(resp.Kind, ShouldEqual, "APIResource")
				})

				Convey("The response should contain a link to the pipeline run uri", func() {
					So(resp.Links, ShouldContainKey, "run")
				})
			})

			Convey("When the method is not GET", func() {
				r, err := http.NewRequest("POST", uri, strings.NewReader(""))
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with 405", func() {
					So(w.Code, ShouldEqual, 405)
				})
			})
		})
	})
}
