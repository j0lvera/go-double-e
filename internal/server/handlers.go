package server

import (
	"encoding/json"
	db "github.com/j0lvera/go-double-e/internal/db/generated"
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
	// Decode the request body into a CreateUserParams struct
	userParams, err := decode[db.CreateUserParams](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the user
	user, err := s.client.Queries.CreateUser(r.Context(), userParams)
	if err != nil {
		// Return an unknown error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// We use a custom response so we don't leak the password hash
	res := struct {
		Id  int64 `json:"id"`
		msg string
	}{
		Id:  user.ID,
		msg: "User created successfully",
	}

	err = encode(w, r, http.StatusCreated, res)
	if err != nil {
		return
	}
}
