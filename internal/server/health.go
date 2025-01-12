package server

import (
	_ "github.com/j0lvera/go-double-e/internal/db"
	"log/slog"
	"net/http"
)

// HandleHealthCheck is a simple health check handler
func (s *Server) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	res := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	// TODO: implement .Ping() on the Client to check the connection
	//if err := s.client.DB().Ping(r.Context()); err != nil {
	//	health.Status = "error"
	//	w.WriteHeader(http.StatusServiceUnavailable)
	//}

	err := WriteResponse(w, http.StatusOK, res)
	if err != nil {
		return
	}
	slog.Info("Health check")
}
