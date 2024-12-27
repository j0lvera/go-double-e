package server

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/j0lvera/go-double-e/internal/db"
	dbGen "github.com/j0lvera/go-double-e/internal/db/generated"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"os"
	"time"
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

	err := writeResponse(w, http.StatusOK, res)
	if err != nil {
		return
	}
	slog.Info("Health check")
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Only validation errors should be returned to the client
	// all other errors should be logged and handled internally

	// Decode the request body into a CreateUserParams struct
	req, err := decode[CreateUserRequest](r)
	if err != nil {
		slog.Info("Failed to decode request body", "error", err)
		writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// Validation
	// https://github.com/go-playground/validator?tab=readme-ov-file#special-notes
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(req); err != nil {
		validationErrors := ParseValidationErrors(err)
		slog.Info("Validation error", "errors", validationErrors)

		// Return an array of validation errors, e.g.:
		// {
		//   "errors": [
		//     {
		//       "field": "email",
		//       "message": "This field is required"
		//     },
		//     {
		//       "field": "password",
		//       "message": "This field must be at least 8 characters long"
		//     }
		//   ]
		// }
		res := map[string][]ValidationError{
			"errors": validationErrors,
		}
		writeError(w, res, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		http.Error(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	userParams := dbGen.CreateUserParams{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	user, err := s.client.Queries.CreateUser(r.Context(), userParams)
	if err != nil {
		slog.Debug("received error from database", "error", err, "error_type", fmt.Sprintf("%T", err))

		if dbErr := db.ParseDBError(err); dbErr != nil {
			switch dbErr.Constraint {
			// TODO: should we let the user know what this error is?
			// If i remember correctly, letting people know when the email already exists is a security concern.

			// This is the only error we want to give to the user (maybe).
			case "users_email_unique":
				slog.Info("duplicate email attempt",
					"constraint", dbErr.Constraint,
					"field", dbErr.Field,
					"email", userParams.Email,
					"message", dbErr.Message,
				)

				writeError(w, ErrUserAlreadyExists, http.StatusConflict)
				return
			default:
				slog.Info("database constraint error",
					"constraint", dbErr.Constraint,
					"field", dbErr.Field,
					"email", userParams.Email,
					"message", dbErr.Message,
				)
				writeError(w, ErrInternalServerError, http.StatusInternalServerError)
				return
			}
		}

		slog.Error("Failed to create user", "error", err)
		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
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

	err = writeResponse(w, http.StatusCreated, res)
	if err != nil {
		return
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Claims struct {
	//Email string `json:"email"`
	Uuid string `json:"uuid"`
	jwt.RegisteredClaims
}

func (s *Server) HandleLoginUser(w http.ResponseWriter, r *http.Request) {
	req, err := decode[LoginRequest](r)
	if err != nil {
		slog.Info("Failed to decode request body", "error", err)
		writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(req); err != nil {
		validationErrors := ParseValidationErrors(err)
		slog.Info("Validation error", "errors", validationErrors)
		res := map[string][]ValidationError{
			"errors": validationErrors,
		}
		writeError(w, res, http.StatusBadRequest)
		return
	}

	user, err := s.client.Queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		slog.Error("Failed to get user from database", "error", err)
		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		slog.Info("Invalid password attempt", "error", err)
		writeError(w, ErrInvalidCredentials, http.StatusUnauthorized)
		return
	}

	claims := Claims{
		Uuid: user.Uuid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			//ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		slog.Error("Failed to sign token", "error", err)
		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    signedToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	res := map[string]string{
		"msg": "Logged in successfully",
	}

	err = writeResponse(w, http.StatusOK, res)
	if err != nil {
		return
	}
}
