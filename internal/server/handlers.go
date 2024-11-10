package server

import (
	"encoding/json"
	"net/http"
)

// HandleHealthCheck is a simple health check handler
func (s *Server) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	// TODO: implement .Ping() on the Client to check the connection
	//if err := s.client.DB().Ping(r.Context()); err != nil {
	//	health.Status = "error"
	//	w.WriteHeader(http.StatusServiceUnavailable)
	//}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := s.client.CreateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		return
	}
}
