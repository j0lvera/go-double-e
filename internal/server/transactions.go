package server

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	dbGen "github.com/j0lvera/go-double-e/internal/db/generated"
	"github.com/jackc/pgx/v5/pgtype"
	"log/slog"
	"net/http"
	"time"
)

type CreateTransactionRequest struct {
	Amount            int64                  `json:"amount" validate:"required"`
	Date              time.Time              `json:"date" validate:"required"`
	Description       string                 `json:"description,omitempty" validate:"max=255"`
	Metadata          map[string]interface{} `json:"metadata"`
	CreditAccountUUID string                 `json:"credit_account_uuid" validate:"required"`
	DebitAccountUUID  string                 `json:"debit_account_uuid" validate:"required"`
	LedgerUUID        string                 `json:"ledger_uuid" validate:"required"`
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

	// cast the date field to a pgtype.Date
	date := pgtype.Date{
		Time:  req.Date,
		Valid: true,
	}

	//// cast the description field to a pgtype.Text
	//description := pgtype.Text{
	//	String: req.Description,
	//	Valid:  true,
	//}

	// marshal the metadata field
	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		slog.Info("unable to marshal metadata", "error", err)
		slog.Debug("metadata marshalling", "metadata", req.Metadata)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	transactionParams := dbGen.CreateTransactionParams{
		Amount:            req.Amount,
		Date:              date,
		Description:       req.Description,
		Metadata:          metadataBytes,
		CreditAccountUuid: req.CreditAccountUUID,
		DebitAccountUuid:  req.DebitAccountUUID,
		LedgerUuid:        req.LedgerUUID,
	}

	transaction, err := s.client.Queries.CreateTransaction(r.Context(), transactionParams)
	if err != nil {
		slog.Error("unable to create transaction", "error", err)
		slog.Debug("transaction creation", "params", transactionParams)

		// TODO:
		// - [ ] handle errors, e.g., "ERROR: Total balance of entries must be 0 (SQLSTATE P0001)" should be invalid request.

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	slog.Debug("transaction creation",
		"transaction", transaction,
		"query_time", time.Since(startReqTime),
	)

	detail := struct {
		UUID        string      `json:"uuid"`
		Amount      int64       `json:"amount"`
		Date        time.Time   `json:"date"`
		Description pgtype.Text `json:"description"`
	}{
		UUID:        transaction.Uuid,
		Amount:      transaction.Amount,
		Date:        transaction.Date.Time,
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
