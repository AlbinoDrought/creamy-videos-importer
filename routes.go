package main

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"
)

type routeDef struct {
	Method  string
	Path    string
	Name    string
	Handler http.HandlerFunc
}

func makeRouter(routes []routeDef) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Path).
			Name(route.Name).
			Handler(route.Handler)
	}

	router.Use(func(handler http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(os.Stdout, handler)
	})

	return router
}
