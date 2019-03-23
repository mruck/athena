package server

import (
	"net/http"

	"github.com/mruck/athena/frontend/database"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() (*mux.Router, error) {
	db := database.MustGetDatabase("localhost", "27017", "athena")
	server, err := NewServer(db)
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter().StrictSlash(true)
	routes := server.getRoutes()
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	return router, nil
}
