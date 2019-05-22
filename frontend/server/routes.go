package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mruck/athena/lib/database"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() (*mux.Router, error) {
	host := MustGetHost()
	db := database.MustGetDatabase(host, "27017", "athena")
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
