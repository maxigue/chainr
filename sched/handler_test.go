package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/Tyrame/chainr/sched/internal/httputil"
)

func TestHandler(t *testing.T) {
	Convey("Scenario: get API resources", t, func() {
		Convey("Given API resources are requested", func() {
			w := httptest.NewRecorder()
			handler := http.Handler(NewHandler())
			uri := "/api"

			Convey("When the method is GET", func() {
				r, err := http.NewRequest("GET", uri, nil)
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)
				var resp httputil.ResponseBody
				json.NewDecoder(w.Body).Decode(&resp)

				Convey("The request should succeed with 200", func() {
					So(w.Code, ShouldEqual, 200)
				})

				Convey("The response should contain APIResourceList", func() {
					So(resp.Kind, ShouldEqual, "APIResourceList")
				})

				Convey("The response should contain links to resources", func() {
					So(resp.Links, ShouldContainKey, "runs")
				})
			})

			Convey("When the method is not GET", func() {
				r, err := http.NewRequest("POST", uri, nil)
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)

				Convey("The request should fail with code 405", func() {
					So(w.Code, ShouldEqual, 405)
				})

				Convey("The response should have the Allow header", func() {
					So(w.Header().Get("Allow"), ShouldEqual, "GET")
				})
			})
		})
	})

	Convey("Scenario: resource not found", t, func() {
		Convey("Given a resource is requested", func() {
			w := httptest.NewRecorder()
			handler := http.Handler(NewHandler())

			Convey("When the resource is not found", func() {
				r, err := http.NewRequest("GET", "/notexisting", nil)
				if err != nil {
					t.Fatal(err)
				}
				handler.ServeHTTP(w, r)
				var resp httputil.ResponseBody
				json.NewDecoder(w.Body).Decode(&resp)

				Convey("The request should fail with code 404", func() {
					So(w.Code, ShouldEqual, 404)
				})

				Convey("The response should contain an Error", func() {
					So(resp.Kind, ShouldEqual, "Error")
				})
			})
		})
	})
}
