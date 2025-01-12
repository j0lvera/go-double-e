package server

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	_ "github.com/j0lvera/go-double-e/internal/db"
	dbGen "github.com/j0lvera/go-double-e/internal/db/generated"
	"github.com/jackc/pgx/v5/pgtype"
	"log/slog"
	"net/http"
	"time"
)

type CreateLedgerRequest struct {
	Name        string                 `json:"name" validate:"required,max=255"`
	Description string                 `json:"description,omitempty" validate:"max=255"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// HandleCreateLedger is the handler for creating a new ledger
func (s *Server) HandleCreateLedger(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()
	slog.Debug("ledger.create.start",
		"method", r.Method,
		"path", r.URL.Path,
		"body", r.Body,
		"remote_add", r.RemoteAddr,
	)

	// Decode the request body
	req, err := Decode[CreateLedgerRequest](r)
	if err != nil {
		slog.Info("unable to Decode request body", "error", err)
		slog.Debug("body decoding", "body", r.Body, "error", err)
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	slog.Debug("body decoding", "request", req)

	// validate the request
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(req); err != nil {
		validationErrors := ParseValidationErrors(err)

		res := map[string][]ValidationError{
			"errors": validationErrors,
		}

		slog.Info("unable to validate request", "error", err)
		slog.Debug("request validation", "body", r.Body, "validation_errors", res)

		WriteError(w, res, http.StatusBadRequest)
		return
	}

	// cast the description field to a pgtype.Text
	description := pgtype.Text{
		String: req.Description,
		Valid:  true,
	}

	// marshal the metadata field
	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		slog.Info("unable to marshal metadata", "error", err)
		slog.Debug("metadata marshalling", "metadata", req.Metadata)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	startQueryTime := time.Now()

	// add the ledger to the database
	ledgerParams := dbGen.CreateLedgerParams{
		Name:        req.Name,
		Description: description,
		Metadata:    metadataBytes,
	}
	ledger, err := s.client.Queries.CreateLedger(r.Context(), ledgerParams)
	if err != nil {
		slog.Error("unable to create ledger", "error", err)
		slog.Debug("ledger creation", "params", ledgerParams, "error", err)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	slog.Debug(
		"ledger creation",
		"uuid", ledger.Uuid,
		"name", ledger.Name,
		"query_time", time.Since(startQueryTime),
	)

	// format the response
	detail := struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	}{
		UUID: ledger.Uuid,
		Name: ledger.Name,
	}

	res := NewResponse("OK", 1, "OBJ", detail)
	err = WriteResponse(w, http.StatusCreated, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res, "error", err)
		return
	}
	slog.Info("ledger created", "ledger_uuid", ledger.Uuid, "name", ledger.Name)
	slog.Debug(
		"ledger.create.complete",
		"ledger_uuid", ledger.Uuid,
		"duration", time.Since(startReqTime),
	)
}

type UpdateLedgerRequest struct {
	Name        string                 `json:"name,omitempty" validate:"max=255"`
	Description string                 `json:"description,omitempty" validate:"max=255"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (s *Server) HandleUpdateLedger(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()

	slog.Debug("ledger.update.start",
		"method", r.Method,
		"path", r.URL.Path,
		"body", r.Body,
		"remote_add", r.RemoteAddr,
	)

	ledgerUUID := r.PathValue("id")
	currentLedger, err := s.client.Queries.GetLedger(r.Context(), ledgerUUID)
	if err != nil {
		slog.Error("unable to get ledger", "error", err)
		slog.Debug("ledger retrieval", "uuid", ledgerUUID)
		WriteError(w, ErrNotFound, http.StatusNotFound)
		return
	}

	// Decode the request body
	req, err := Decode[UpdateLedgerRequest](r)
	if err != nil {
		slog.Info("unable to Decode request body", "error", err)
		slog.Debug("body decoding", "body", r.Body)
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	slog.Debug("body decoding", "request", req)

	// validate the request
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(req); err != nil {
		validationErrors := ParseValidationErrors(err)

		res := map[string][]ValidationError{
			"errors": validationErrors,
		}

		slog.Info("unable to validate request", "error", err)
		slog.Debug("request validation", "body", r.Body)

		WriteError(w, res, http.StatusBadRequest)
		return
	}

	ledgerParams := dbGen.UpdateLedgerParams{
		Uuid: ledgerUUID,
	}

	if req.Name != "" {
		ledgerParams.Name = req.Name
	} else {
		ledgerParams.Name = currentLedger.Name
	}

	if req.Description != "" {
		// cast the description field to a pgtype.Text
		description := pgtype.Text{
			String: req.Description,
			Valid:  true,
		}
		ledgerParams.Description = description
	} else {
		ledgerParams.Description = currentLedger.Description
	}

	if req.Metadata != nil {
		// marshal the metadata field
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			slog.Info("unable to marshal metadata", "error", err)
			slog.Debug("metadata marshalling", "metadata", req.Metadata)

			WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
			return
		}
		ledgerParams.Metadata = metadataBytes
	} else {
		ledgerParams.Metadata = currentLedger.Metadata
	}

	startQueryTime := time.Now()

	ledger, err := s.client.Queries.UpdateLedger(r.Context(), ledgerParams)
	if err != nil {
		slog.Error("unable to update ledger", "error", err)
		slog.Debug("ledger update", "params", ledgerParams, "error", err)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	slog.Debug("ledger update",
		"uuid", ledger.Uuid,
		"name", ledger.Name,
		"query_time", time.Since(startQueryTime),
	)

	// format the response
	detail := struct {
		UUID        string                 `json:"uuid"`
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Metadata    map[string]interface{} `json:"metadata"`
	}{
		UUID:        ledger.Uuid,
		Name:        ledger.Name,
		Description: ledger.Description.String,
		Metadata:    req.Metadata,
	}

	res := NewResponse("OK", 1, "OBJ", detail)
	err = WriteResponse(w, http.StatusOK, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res, "error", err)
		return
	}
	slog.Info("ledger updated", "ledger_uuid", ledger.Uuid, "name", ledger.Name)
	slog.Debug(
		"ledger.update.complete",
		"ledger_uuid", ledger.Uuid,
		"duration", time.Since(startReqTime),
	)
}

type MetadataQuery struct {
	Metadata map[string]any `schema:"metadata"`
}

// HandleListLedgers returns a list of ledgers
// whatever the request body is, it will be ignored.
// you must send a query string parameter in the shape of:
//   - ?metadata[key]=value
//   - ?metadata[key]=value&metadata[key]=value
//
// if the metadata param returns 0 results, the server will return 404.
// if no query string parameter is present, it will return bad request.
func (s *Server) HandleListLedgers(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()
	slog.Debug("ledger.list.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_add", r.RemoteAddr,
	)

	// parse and marshal the metadata field
	metadataBytes, err := parseMetadataParam(r)
	if err != nil {
		slog.Info("unable to parse metadata query param", "error", err)
		slog.Debug("metadata parsing", "raw_query", r.URL.RawQuery)

		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	slog.Debug("metadata parsing", "raw_query", r.URL.RawQuery, "metadata", string(metadataBytes))

	startQueryTime := time.Now()

	// get the ledgers from the database
	ledgers, err := s.client.Queries.ListLedgers(r.Context(), metadataBytes)
	if err != nil {
		slog.Error("unable to list ledgers", "error", err)

		deadline, _ := r.Context().Deadline()
		slog.Debug(
			"database querying",
			"query_timeout", deadline,
			"metadata_filter", string(metadataBytes),
		)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	ledgersCount := len(ledgers)

	slog.Debug(
		"database querying",
		"ledgers_count", ledgersCount,
		"metadata_filter", string(metadataBytes),
		"query_time", time.Since(startQueryTime),
	)

	// no ledgers found, return 404
	if ledgersCount == 0 {
		slog.Info("unable to find queries", "metadata_filter", string(metadataBytes))
		slog.Debug(
			"ledger.list.complete",
			"ledgers_count", ledgersCount,
			"duration", time.Since(startReqTime),
		)

		WriteError(w, ErrNotFound, http.StatusNotFound)
		return
	}

	// format the response
	detail := ledgers

	slog.Debug(
		"response preparation",
		"ledgers_count", ledgersCount,
		"first_ledger_uuid", ledgers[0].Uuid,
	)

	res := NewResponse("OK", ledgersCount, "LIST", detail)
	err = WriteResponse(w, http.StatusOK, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res)
		return
	}
	slog.Info("ledgers listed", "count", ledgersCount)
	slog.Debug(
		"ledger.list.complete",
		"ledgers_count", ledgersCount,
		"duration", time.Since(startReqTime),
	)
}
