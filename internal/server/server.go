package server

import (
	"github.com/j0lvera/go-double-e/pkg/doublee"
	"net/http"
)

type Server struct {
	client *doublee.Client
}

func NewServer(client *doublee.Client) http.Handler {
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
	mux.HandleFunc("/health", s.HandleHealthCheck)
}
