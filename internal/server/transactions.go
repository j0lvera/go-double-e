package server

import (
	"context"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	dbGen "github.com/j0lvera/go-double-e/internal/db/generated"
	"github.com/j0lvera/opi/crud"
	"github.com/jackc/pgx/v5/pgtype"
	"log/slog"
	"net/http"
	"time"
)

type Lister struct {
	queries *dbGen.Queries
}

type ListQuery struct {
	LedgerUUID string `form:"ledger_uuid" json:"ledger_uuid" validate:"required"`
	Metadata   string `form:"metadata" json:"metadata"`
	crud.PaginatedQuery
}

func (q ListQuery) GetPagination() crud.PaginatedQuery {
	return q.PaginatedQuery
}

func (l *Lister) List(ctx context.Context, query ListQuery, offset int, limit int) ([]*dbGen.ListTransactionsRow, error) {
	params := dbGen.ListTransactionsParams{
		Limit:      int32(limit),
		Offset:     int32(offset),
		Metadata:   []byte(query.Metadata),
		LedgerUuid: query.LedgerUUID,
	}
	return l.queries.ListTransactions(ctx, params)
}

func (l *Lister) Count(ctx context.Context, query ListQuery) (int64, error) {
	params := dbGen.GetTransactionsCountParams{
		LedgerUuid: query.LedgerUUID,
		Metadata:   []byte(query.Metadata),
	}
	return l.queries.GetTransactionsCount(ctx, params)
}

func NewListHandler(queries *dbGen.Queries) *crud.ListHandler[*dbGen.ListTransactionsRow, ListQuery] {
	return crud.NewListHandler[*dbGen.ListTransactionsRow, ListQuery](
		&Lister{queries: queries},
		&crud.DefaultResponseWriter{},
	)
}

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

type UpdateTransactionRequest struct {
	Amount            *int64                  `json:"amount,omitempty"`
	Date              *time.Time              `json:"date,omitempty"`
	Description       *pgtype.Text            `json:"description,omitempty"`
	Metadata          *map[string]interface{} `json:"metadata,omitempty"`
	CreditAccountUuid *pgtype.Text            `json:"credit_account_uuid,omitempty"`
	DebitAccountUuid  *pgtype.Text            `json:"debit_account_uuid,omitempty"`
	LedgerID          *pgtype.Text            `json:"ledger_uuid,omitempty"`
}

func (s *Server) HandleUpdateTransaction(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()

	slog.Debug("transaction.update.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_add", r.RemoteAddr,
	)

	txnUUID := r.PathValue("uuid")

	// decode the request body
	req, err := Decode[UpdateTransactionRequest](r)
	if err != nil {
		slog.Info("unable to decode request body", "error", err)
		slog.Debug("body decoding", "body", r.Body)
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	slog.Debug("body decoding", "request", req)

	// return invalid request on empty body, e.g., {}
	if isEmptyJSON(req) {
		slog.Info("empty update request body")
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// validate the request
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err = validate.Struct(req); err != nil {
		validationErrors := ParseValidationErrors(err)

		res := map[string][]ValidationError{
			"errors": validationErrors,
		}

		slog.Info("unable to validate request", "errors", res)
		slog.Debug("request validation", "body", r.Body)

		WriteError(w, res, http.StatusBadRequest)
		return
	}

	// marshal the metadata field
	metadataValue, metadataIsValid := ptrValue(req.Metadata)
	var metadataBytes []byte
	if metadataIsValid {
		metadataBytes, err = json.Marshal(metadataValue)
		if err != nil {
			slog.Info("unable to marshal metadata", "error", err)
			slog.Debug("metadata marshaling", "metadata", metadataValue)
			WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
			return
		}
	}

	// cast the amount to pgtype.Int8
	amountValue, amountIsValid := ptrValue(req.Amount)
	amount := pgtype.Int8{
		Int64: amountValue,
		Valid: amountIsValid,
	}

	// cast the date to pgtype.Date
	dateValue, dateIsValid := ptrValue(req.Date)
	date := pgtype.Date{
		Time:  dateValue,
		Valid: dateIsValid,
	}

	// cast the description to pgtype.Text
	descValue, descIsValid := ptrValue(req.Description)
	description := pgtype.Text{
		String: descValue.String,
		Valid:  descIsValid,
	}

	// cast the credit account id to pgtype.int8
	creditIdValue, creditIdIsValid := ptrValue(req.CreditAccountUuid)
	creditAccountUuid := pgtype.Text{
		String: creditIdValue.String,
		Valid:  creditIdIsValid,
	}

	// cast the debit account id to pgtype.int8
	debitIdValue, debitIdIsValid := ptrValue(req.DebitAccountUuid)
	debitAccountUuid := pgtype.Text{
		String: debitIdValue.String,
		Valid:  debitIdIsValid,
	}

	// cast the ledger id to pgtype.int8
	ledgerIdValue, ledgerIdIsValid := ptrValue(req.LedgerID)
	ledgerUuid := pgtype.Text{
		String: ledgerIdValue.String,
		Valid:  ledgerIdIsValid,
	}

	txnParams := dbGen.UpdateTransactionParams{
		Uuid:              txnUUID,
		Amount:            amount,
		Date:              date,
		Description:       description,
		Metadata:          metadataBytes,
		CreditAccountUuid: creditAccountUuid,
		DebitAccountUuid:  debitAccountUuid,
		LedgerUuid:        ledgerUuid,
	}

	startQueryTime := time.Now()

	txn, err := s.client.Queries.UpdateTransaction(r.Context(), txnParams)
	if err != nil {
		slog.Error("unable to update transaction", "error", err)
		slog.Debug("transaction update", "params", txnParams)
		deadline, _ := r.Context().Deadline()
		slog.Debug("transaction update", "query_timeout", deadline)
		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	slog.Debug("transaction update",
		"uuid", txn.Uuid,
		"query_time", time.Since(startQueryTime),
	)

	// format response
	detail := struct {
		Uuid        string      `json:"uuid"`
		Amount      int64       `json:"amount"`
		Date        pgtype.Date `json:"date"`
		Description pgtype.Text `json:"description"`
		//Metadata map[string]interface{} `json:"metadata"`
	}{
		Uuid:        txn.Uuid,
		Amount:      txn.Amount,
		Date:        txn.Date,
		Description: txn.Description,
		//Metadata: txn.Metadata,
	}

	res := NewResponse("OK", 1, "OBJ", detail)
	err = WriteResponse(w, http.StatusOK, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res)
		return
	}
	slog.Info("transaction updated", "uuid", txn.Uuid)
	slog.Debug(
		"transaction.update.complete",
		"uuid", txn.Uuid,
		"duration", time.Since(startReqTime),
	)
}

func (s *Server) HandleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()
	slog.Debug("transaction.delete.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_add", r.RemoteAddr,
	)

	txnUUID := r.PathValue("uuid")
	if txnUUID == "" {
		slog.Info("transaction uuid is required")
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	startQueryTime := time.Now()

	err := s.client.Queries.DeleteTransaction(r.Context(), txnUUID)
	if err != nil {
		slog.Error("unable to delete transaction", "error", err)
		slog.Debug("transaction deletion", "uuid", txnUUID)

		deadline, _ := r.Context().Deadline()
		slog.Debug("transaction deletion", "query_timeout", deadline)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	slog.Debug("transaction deletion",
		"uuid", txnUUID,
		"query_time", time.Since(startQueryTime),
	)

	res := NewResponse("OK", 1, "OBJ", nil)
	err = WriteResponse(w, http.StatusNoContent, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res)
		return
	}

	slog.Info("transaction deleted", "uuid", txnUUID)
	slog.Debug(
		"transaction.delete.complete",
		"uuid", txnUUID,
		"duration", time.Since(startReqTime),
	)
}
