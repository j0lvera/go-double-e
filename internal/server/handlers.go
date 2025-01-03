package server

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	_ "github.com/j0lvera/go-double-e/internal/db"
	dbGen "github.com/j0lvera/go-double-e/internal/db/generated"
	"github.com/jackc/pgx/v5/pgtype"
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
	// decode the request body
	req, err := decode[CreateLedgerRequest](r)
	if err != nil {
		slog.Info("Failed to decode request", "error", err)
		writeError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// validate the request
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(req); err != nil {
		validationErrors := ParseValidationErrors(err)
		slog.Info("Failed to validate request", "error", err)

		res := map[string][]ValidationError{
			"errors": validationErrors,
		}
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
		slog.Debug("Failed to marshal metadata field", "error", err)
		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// add the ledger to the database
	ledgerParams := dbGen.CreateLedgerParams{
		Name:        req.Name,
		Description: description,
		Metadata:    metadataBytes,
	}
	ledger, err := s.client.Queries.CreateLedger(r.Context(), ledgerParams)
	if err != nil {
		slog.Debug("Failed to create ledger", "error", err)
		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

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
		return
	}
	slog.Info("Ledger created", "uuid", ledger.Uuid)
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
// if the metadata param is not present, it will be ignored and the server will return 404.
// if no query string parameter is present, it will return bad request.
func (s *Server) HandleListLedgers(w http.ResponseWriter, r *http.Request) {
	// parse and marshal the metadata field
	metadataBytes, err := parseMetadataParam(r)
	if err != nil {
		slog.Debug("Failed to parse and marshal metadata field", "error", err)
		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// get the ledgers from the database
	ledgers, err := s.client.Queries.ListLedgers(r.Context(), metadataBytes)
	if err != nil {
		slog.Debug("Failed to list ledgers", "error", err)
		writeError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	ledgersCount := len(ledgers)

	// no ledgers found, return 404
	if ledgersCount == 0 {
		writeError(w, ErrNotFound, http.StatusNotFound)
		return
	}

	// format the response
	detail := struct {
		Ledgers []dbGen.Ledger `json:"ledgers"`
	}{
		Ledgers: ledgers,
	}

	res := NewResponse("OK", ledgersCount, "LIST", detail)
	err = writeResponse(w, http.StatusOK, res)
	if err != nil {
		return
	}
	slog.Info("Listed ledgers", "count", ledgersCount)
}
