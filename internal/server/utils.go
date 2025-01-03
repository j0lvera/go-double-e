package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func writeResponse[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("writeResponse json: %w", err)
	}

	return nil
}

// query params decoder
func decodeParams[T any](r *http.Request) (T, error) {
	var v T

	if err := schema.NewDecoder().Decode(&v, r.URL.Query()); err != nil {
		return v, fmt.Errorf("decode query params: %w", err)
	}

	return v, nil
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

func parseMetadataParam(r *http.Request) ([]byte, error) {
	queryValues := r.URL.Query()
	prefix := "metadata."

	metadataFilter := make(map[string]interface{})
	for key, value := range queryValues {
		if strings.HasPrefix(key, prefix) {
			actualKey := strings.TrimPrefix(key, prefix)
			// try to convert to number if possible
			if num, err := strconv.ParseInt(value[0], 10, 64); err == nil {
				metadataFilter[actualKey] = num
			} else {
				metadataFilter[actualKey] = value[0]
			}
		}
	}

	if len(metadataFilter) == 0 {
		return nil, fmt.Errorf("no metadata parameters found")
	}

	// convert to JSON for database query
	metadataBytes, err := json.Marshal(metadataFilter)
	if err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}

	return metadataBytes, nil
}
