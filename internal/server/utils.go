package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func WriteResponse[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("WriteResponse json: %w", err)
	}

	return nil
}

// query params decoder
func decodeParams[T any](r *http.Request) (T, error) {
	var v T

	if err := schema.NewDecoder().Decode(&v, r.URL.Query()); err != nil {
		return v, fmt.Errorf("Decode query params: %w", err)
	}

	return v, nil
}

func Decode[T any](r *http.Request) (T, error) {
	var v T

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("Decode json: %w", err)
	}

	return v, nil
}

func WriteError(w http.ResponseWriter, message interface{}, status int) {
	res := ErrorResponse{
		Status:  status,
		Message: message,
	}

	err := WriteResponse(w, status, res)
	if err != nil {
		// If we fail to WriteResponse the error response, we log the error and return a generic 500 Internal Server Error
		slog.Error("Failed to WriteResponse error response", "error", err)
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

// MapNonZeroFields
// these two interface might have a different amount of fields, so we need to loop through
// the one that has the most.
// requirements:
// - they have to be the same shape.
// - each field should be of the same type.
func MapNonZeroFields[T any, U any](src *T, dest *U) {
	srcVal := reflect.ValueOf(src).Elem()
	destVal := reflect.ValueOf(dest).Elem()
	srcType := srcVal.Type()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldName := srcType.Field(i).Name
		destField := destVal.FieldByName(srcFieldName)

		if !srcField.IsZero() && destField.IsValid() && destField.CanSet() {
			if srcField.Type() == destField.Type() {
				destField.Set(srcField)
			}
		}
	}
}

// ptrValue returns the value of a pointer if
// it is not nil, otherwise it returns the zero
// value of the type and false.
func ptrValue[T any](p *T) (v T, valid bool) {
	if p == nil {
		return v, false
	}
	return *p, true
}

// isEmptyJSON checks if a struct is empty by checking if all of its
// fields are zero values.
func isEmptyJSON(v interface{}) bool {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	nonNilFields := 0
	for i := 0; i < val.NumField(); i++ {
		if !val.Field(i).IsZero() {
			nonNilFields++
		}
	}

	return nonNilFields == 0
}
