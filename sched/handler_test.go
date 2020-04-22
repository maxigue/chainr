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
				var list apiResourceList
				json.NewDecoder(w.Body).Decode(&list)

				Convey("The request should succeed with 200", func() {
					So(w.Code, ShouldEqual, 200)
				})

				Convey("The response should have the Content-Type application/json", func() {
					So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				})

				Convey("The response should be of kind APIResourceList", func() {
					So(list.Kind, ShouldEqual, "APIResourceList")
				})

				Convey("The response should have a link to itself", func() {
					So(list.Metadata.SelfLink, ShouldEqual, "/api")
				})

				Convey("The response should contain resources", func() {
					So(list.Resources, ShouldContainKey, "runs")
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

				Convey("The response should have the Content-Type application/json", func() {
					So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
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
				var error httputil.Error
				json.NewDecoder(w.Body).Decode(&error)

				Convey("The request should fail with code 404", func() {
					So(w.Code, ShouldEqual, 404)
				})

				Convey("The response should be of kind Error", func() {
					So(error.Kind, ShouldEqual, "Error")
				})
			})
		})
	})
}
