package server

import (
	"github.com/go-playground/validator/v10"
	db "github.com/j0lvera/go-double-e/internal/db/generated"
	"golang.org/x/crypto/bcrypt"
	"log/slog"

	//db "github.com/j0lvera/go-double-e/internal/db/generated"
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

	err := encode(w, r, http.StatusOK, res)
	if err != nil {
		return
	}
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a CreateUserParams struct
	req, err := decode[CreateUserRequest](r)
	if err != nil {
		slog.Info("Failed to decode request body", "error", err)
		http.Error(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// Validation
	// https://github.com/go-playground/validator?tab=readme-ov-file#special-notes
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(req); err != nil {
		//fmt.Print(map[string]string{"error": err.Error()})
		slog.Info("Validation error", "error", err)
		http.Error(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		http.Error(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	userParams := db.CreateUserParams{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	user, err := s.client.Queries.CreateUser(r.Context(), userParams)
	if err != nil {
		slog.Error("Failed to create user", "error", err)
		http.Error(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// We use a custom response so we don't leak the password hash
	res := struct {
		Uuid string `json:"uuid"`
		msg  string
	}{
		Uuid: user.Uuid,
		msg:  "User created successfully",
	}

	err = encode(w, r, http.StatusCreated, res)
	if err != nil {
		return
	}
}
