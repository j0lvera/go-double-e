package server

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	dbGen "github.com/j0lvera/go-double-e/internal/db/generated"
	"log/slog"
	"net/http"
	"time"
)

type CreateAccountRequest struct {
	Name       string                 `json:"name" validate:"required,max=255"`
	Type       string                 `json:"type" validate:"required"`
	Metadata   map[string]interface{} `json:"metadata"`
	LedgerUUID string                 `json:"ledger_uuid" validate:"required"`
}

func (s *Server) HandleCreateAccount(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()
	slog.Debug("account.create.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remoted_add", r.RemoteAddr,
	)

	// decode the request body
	req, err := Decode[CreateAccountRequest](r)
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

	// cast the account type to the db type
	accountType := dbGen.AccountType(req.Type)

	// marshal the metadata field
	metadataByes, err := json.Marshal(req.Metadata)
	if err != nil {
		slog.Info("unable to marshal metadata", "error", err)
		slog.Debug("metadata marshalling", "metadata", req.Metadata)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	accountParams := dbGen.CreateAccountParams{
		Name:       req.Name,
		Type:       accountType,
		Metadata:   metadataByes,
		LedgerUuid: req.LedgerUUID,
	}
	account, err := s.client.Queries.CreateAccount(r.Context(), accountParams)
	if err != nil {
		slog.Error("unable to create account", "error", err)
		slog.Debug("account creation", "params", accountParams)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	slog.Debug("account creation",
		"uuid", account.Uuid,
		"name", account.Name,
		"query_time", time.Since(startReqTime),
	)

	detail := struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	}{
		UUID: account.Uuid,
		Name: account.Name,
	}

	res := NewResponse("OK", 1, "OBJ", detail)
	err = WriteResponse(w, http.StatusCreated, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res)
		return
	}
	slog.Info("account created", "uuid", account.Uuid, "name", account.Name)
	slog.Debug("account.create.complete",
		"account_uuid", account.Uuid,
		"duration", time.Since(startReqTime),
	)
}

func (s *Server) HandleListAccounts(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()
	slog.Debug("account.list.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_add", r.RemoteAddr,
	)

	// TODO:
	// - [ ] create a validation where we create a struct from the query params
	// ledger_uuid and metadata and we show a validation error instead of an empty
	// invalid request

	queryValues := r.URL.Query()
	ledgerUUID := queryValues.Get("ledger_uuid")
	if ledgerUUID == "" {
		slog.Info("ledger_uuid is required", "query_params", queryValues)
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	metadataBytes, err := parseMetadataParam(r)
	if err != nil {
		slog.Info("unable to parse metadata query param", "error", err)
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	getAccountParams := dbGen.ListAccountsParams{
		LedgerUuid: ledgerUUID,
		Metadata:   metadataBytes,
	}

	startQueryTime := time.Now()

	// get the accounts from the database
	accounts, err := s.client.Queries.ListAccounts(r.Context(), getAccountParams)
	if err != nil {
		slog.Error("unable to list accounts", "error", err)

		deadline, _ := r.Context().Deadline()
		slog.Debug("account listing", "query_timeout", deadline)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	accountsCount := len(accounts)

	slog.Debug("account listing",
		"accounts_count", accountsCount,
		"query_time", time.Since(startQueryTime),
	)

	if accountsCount == 0 {
		slog.Info("no accounts found", "metadata_filter", string(metadataBytes))
		slog.Debug("account.list.complete",
			"accounts_count", accountsCount,
			"duration", time.Since(startReqTime),
		)

		WriteError(w, ErrNotFound, http.StatusNotFound)
		return
	}

	res := NewResponse("OK", accountsCount, "LIST", accounts)
	err = WriteResponse(w, http.StatusOK, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res)
		return
	}

	slog.Info("accounts listed", "count", accountsCount)
	slog.Debug(
		"account.list.complete",
		"accounts_count", accountsCount,
		"duration", time.Since(startReqTime),
	)
}

type UpdateAccountRequest struct {
	Name     string                 `json:"name,omitempty" validate:"max=255"`
	Type     string                 `json:"type,omitempty" validate:""`
	Metadata map[string]interface{} `json:"metadata,omitempty" validate:""`
}

func (s *Server) HandleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	startReqTime := time.Now()

	slog.Debug("account.update.start",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_add", r.RemoteAddr,
	)

	accountUUID := r.PathValue("id")
	currentAccount, err := s.client.Queries.GetAccount(r.Context(), accountUUID)
	if err != nil {
		slog.Error("unable to get account", "error", err)
		slog.Debug("account retrieval", "uuid", accountUUID)
		WriteError(w, ErrNotFound, http.StatusNotFound)
		return
	}

	// decode the request body
	req, err := Decode[UpdateAccountRequest](r)
	if err != nil {
		slog.Info("unable to decode request body", "error", err)
		slog.Debug("body decoding", "body", r.Body)
		WriteError(w, ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	slog.Debug("body decoding", "request", req)

	// TODO:
	// - [ ] as part of the validation, we should check if the account type is valid

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

	accountParams := dbGen.UpdateAccountParams{
		Uuid: accountUUID,
	}

	if req.Name != "" {
		accountParams.Name = req.Name
	} else {
		accountParams.Name = currentAccount.Name
	}

	if req.Type != "" {
		accountParams.Type = dbGen.AccountType(req.Type)
	} else {
		accountParams.Type = currentAccount.Type
	}

	if req.Metadata != nil {
		// marshal the metadata field
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			slog.Info("unable to marshal metadata", "error", err)
			slog.Debug("metadata marshaling", "metadata", req.Metadata)

			WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
			return
		}
		accountParams.Metadata = metadataBytes
	} else {
		accountParams.Metadata = currentAccount.Metadata
	}

	startQueryTime := time.Now()

	account, err := s.client.Queries.UpdateAccount(r.Context(), accountParams)
	if err != nil {
		slog.Error("unable to update account", "error", err)
		slog.Debug("account update", "uuid", accountUUID, "params", accountParams)

		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	slog.Debug("account update",
		"uuid", account.Uuid,
		"name", account.Name,
		"query_time", time.Since(startQueryTime),
	)

	// Unmarshal the metadata for response
	var metadata map[string]interface{}
	if err := json.Unmarshal(account.Metadata, &metadata); err != nil {
		slog.Error("unable to unmarshal metadata", "error", err)
		slog.Debug("metadata unmarshalling", "metadata", account.Metadata)
		WriteError(w, ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// format response
	detail := struct {
		UUID     string                 `json:"uuid"`
		Name     string                 `json:"name"`
		Type     string                 `json:"type"`
		Metadata map[string]interface{} `json:"metadata"`
	}{
		UUID:     account.Uuid,
		Name:     account.Name,
		Type:     string(account.Type),
		Metadata: metadata,
	}

	res := NewResponse("OK", 1, "OBJ", detail)
	err = WriteResponse(w, http.StatusOK, res)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		slog.Debug("response writing", "response", res)
		return
	}

	slog.Info("account updated",
		"account_uuid", account.Uuid,
		"name", account.Name,
	)
	slog.Debug(
		"account.update.complete",
		"account_uuid", account.Uuid,
		"duration", time.Since(startReqTime),
	)
}
