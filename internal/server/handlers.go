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
		"remote_addr", r.RemoteAddr)

	// decode the request body
	req, err := decode[CreateLedgerRequest](r)
	if err != nil {
		slog.Info("unable to decode request body", "error", err)
		slog.Debug("body decoding", "body", r.Body, "error", err)
		writeError(w, ErrInvalidRequest, http.StatusBadRequest)
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
		slog.Debug("request validation", "body", r.Body, "validation_errors", res, "error", err)

		writeError(w, res, http.StatusBadRequest)
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
		slog.Debug("metadata marshalling", "metadata", req.Metadata, "error", err)

		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
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

		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
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
	err = writeResponse(w, http.StatusCreated, res)
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

type MetadataQuery struct {
	Metadata map[string]any `schema:"metadata"`
}

// HandleListLedgers returns a list of ledgers
// whatever the request body is, it will be ignored.
// you must send a query string parameter in the shape of:
//   - ?metadata[key]=value
//   - ?metadata[key]=value&metadata[key]=value
//
// if the metadata param return 0 results, the server will return 404.
// if no query string parameter is present, it will return bad request.
func (s *Server) HandleListLedgers(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()
	slog.Debug("ledger.list.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remoted_addr", r.RemoteAddr)

	// parse and marshal the metadata field
	metadataBytes, err := parseMetadataParam(r)
	if err != nil {
		slog.Info("unable to parse metadata query param", "error", err)
		slog.Debug("metadata parsing", "raw_query", r.URL.RawQuery)

		writeError(w, ErrInvalidRequest, http.StatusBadRequest)
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

		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
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

		writeError(w, ErrNotFound, http.StatusNotFound)
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
	err = writeResponse(w, http.StatusOK, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		return
	}

	slog.Info("ledgers listed", "count", ledgersCount)
	slog.Debug(
		"ledger.list.complete",
		"ledgers_count", ledgersCount,
		"duration", time.Since(startReqTime),
	)
}
