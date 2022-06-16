package main

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

var routes = []string{
	"/",
	"/login",
	"/logout",
	"/register",
	"/members/plans",
	"/members/subscribe",
}

func Test_routes_exist(t *testing.T) {
	mux, ok := testApp.routes().(*chi.Mux)

	if !ok {
		t.Error("routes does not return a chi mux")
	}

	for _, thisRoute := range routes {
		found := false
		chi.Walk(mux, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			if thisRoute == route {
				found = true
			}
			return nil
		})

		if !found {
			t.Errorf("Route %s not found", thisRoute)
		}
	}

}
