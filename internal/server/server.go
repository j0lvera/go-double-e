package server

import (
	"github.com/j0lvera/go-double-e/internal/db"
	"net/http"
)

type Server struct {
	client *db.Client
}

func NewServer(client *db.Client) http.Handler {
	// top level HTTP that applies to all routes, e.g.,
	// CORS, auth middlewares, logging, etc.

	srv := Server{
		client: client,
	}

	mux := http.NewServeMux()
	srv.addRoutes(mux)

	// add middlewares here

	var handler http.Handler = mux
	return handler
}

func (s *Server) addRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.HandleHealthCheck)

	mux.HandleFunc("POST /ledgers", s.HandleCreateLedger)
	mux.HandleFunc("GET /ledgers", s.HandleListLedgers)
}
