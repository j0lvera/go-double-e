package server

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	dbGen "github.com/j0lvera/go-double-e/internal/db/generated"
	"log/slog"
	"net/http"
	"time"
)

type CreateTransactionRequest struct {
	Description string                 `json:"description" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata"`
	LedgerUUID  string                 `json:"ledger_uuid" validate:"required"`
	Entries     []json.RawMessage      `json:"entries" validate:"required"`
}

func (s *Server) HandleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()
	slog.Debug("transaction.create.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remoted_add", r.RemoteAddr,
	)

	// decode the request body
	req, err := Decode[CreateTransactionRequest](r)
	if err != nil {
		slog.Info("unable to decode request body", "error", err)
		slog.Debug("request body decoding", "body", r.Body)
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	slog.Debug("body decoding", "body", r.Body, "request", req)

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

	// marshal the metadata field
	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		slog.Info("unable to marshal metadata", "error", err)
		slog.Debug("metadata marshalling", "metadata", req.Metadata)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// Convert entries to [][]byte
	entriesBytes := make([][]byte, len(req.Entries))
	for i, entry := range req.Entries {
		entriesBytes[i] = []byte(entry)
	}

	transactionParams := dbGen.CreateTransactionWithEntriesParams{
		Description: req.Description,
		Metadata:    metadataBytes,
		LedgerUuid:  req.LedgerUUID,
		Entries:     entriesBytes,
	}

	response, err := s.client.Queries.CreateTransactionWithEntries(r.Context(), transactionParams)
	if err != nil {
		slog.Error("unable to create transaction", "error", err)
		slog.Debug("transaction creation", "params", transactionParams)

		// TODO:
		// - [ ] handle errors, e.g., "ERROR: Total balance of entries must be 0 (SQLSTATE P0001)" should be invalid request.

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	slog.Debug("transaction creation",
		"transaction", response,
		"query_time", time.Since(startReqTime),
	)

	// we cast the type because sqlc doesn't recognize the result type.
	// this issue might be caused because we are using a postgres function.
	type TransactionResponse struct {
		Id          int64  `json:"id"`
		Uuid        string `json:"uuid"`
		Description string `json:"description"`
	}
	responseArr := response.([]interface{})
	transaction := TransactionResponse{
		Id:          responseArr[0].(int64),
		Uuid:        responseArr[1].(string),
		Description: responseArr[2].(string),
	}

	detail := struct {
		UUID        string `json:"uuid"`
		Description string `json:"description"`
	}{
		UUID:        transaction.Uuid,
		Description: transaction.Description,
	}

	res := NewResponse("OK", 1, "OBJ", detail)
	err = WriteResponse(w, http.StatusCreated, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res)
		return
	}

	slog.Info("transaction created", "uuid", transaction.Uuid, "description", transaction.Description)
	slog.Debug("transaction.create.complete",
		"transaction_uuid", transaction.Uuid,
		"duration", time.Since(startReqTime),
	)
}
