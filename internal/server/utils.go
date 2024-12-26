package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func writeResponse[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("writeResponse json: %w", err)
	}

	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	return v, nil
}

func writeError(w http.ResponseWriter, message interface{}, status int) {
	res := ErrorResponse{
		Status:  status,
		Message: message,
	}

	err := writeResponse(w, status, res)
	if err != nil {
		// If we fail to writeResponse the error response, we log the error and return a generic 500 Internal Server Error
		slog.Error("Failed to writeResponse error response", "error", err)
		http.Error(w, ErrInternalServerError, http.StatusInternalServerError)
	}
}
