// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: accounts.sql

package db

import (
	"context"
)

const createAccount = `-- name: CreateAccount :one
     with ledger as (select id
                       from ledgers
                      where uuid = $4::text)
   insert
     into accounts (name, type, metadata, ledger_id)
   values ($1, $2, $3, (select id from ledger))
returning id, uuid, created_at, updated_at, name, type, metadata, ledger_id
`

type CreateAccountParams struct {
	Name       string
	Type       AccountType
	Metadata   []byte
	LedgerUuid string
}

// CreateAccount
//
//	     with ledger as (select id
//	                       from ledgers
//	                      where uuid = $4::text)
//	   insert
//	     into accounts (name, type, metadata, ledger_id)
//	   values ($1, $2, $3, (select id from ledger))
//	returning id, uuid, created_at, updated_at, name, type, metadata, ledger_id
func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRow(ctx, createAccount,
		arg.Name,
		arg.Type,
		arg.Metadata,
		arg.LedgerUuid,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Type,
		&i.Metadata,
		&i.LedgerID,
	)
	return i, err
}

const getAccount = `-- name: GetAccount :one
select id, uuid, created_at, updated_at, name, type, metadata, ledger_id
  from accounts
 where uuid = $1
 limit 1
`

// GetAccount
//
//	select id, uuid, created_at, updated_at, name, type, metadata, ledger_id
//	  from accounts
//	 where uuid = $1
//	 limit 1
func (q *Queries) GetAccount(ctx context.Context, uuid string) (Account, error) {
	row := q.db.QueryRow(ctx, getAccount, uuid)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Type,
		&i.Metadata,
		&i.LedgerID,
	)
	return i, err
}

const listAccounts = `-- name: ListAccounts :many
  with ledger as (select id from ledgers where uuid = $2::text)
select uuid, name, type, metadata
  from accounts
 where ledger_id = (select id from ledger)
   and metadata @> $1::jsonb
`

type ListAccountsParams struct {
	Metadata   []byte
	LedgerUuid string
}

type ListAccountsRow struct {
	Uuid     string
	Name     string
	Type     AccountType
	Metadata []byte
}

// ListAccounts
//
//	  with ledger as (select id from ledgers where uuid = $2::text)
//	select uuid, name, type, metadata
//	  from accounts
//	 where ledger_id = (select id from ledger)
//	   and metadata @> $1::jsonb
func (q *Queries) ListAccounts(ctx context.Context, arg ListAccountsParams) ([]ListAccountsRow, error) {
	rows, err := q.db.Query(ctx, listAccounts, arg.Metadata, arg.LedgerUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAccountsRow
	for rows.Next() {
		var i ListAccountsRow
		if err := rows.Scan(
			&i.Uuid,
			&i.Name,
			&i.Type,
			&i.Metadata,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateAccount = `-- name: UpdateAccount :one
   update accounts
      set name     = coalesce($2, name),
          type     = coalesce($3, type),
          metadata = coalesce($4, metadata)
    where uuid = $1
returning id, uuid, created_at, updated_at, name, type, metadata, ledger_id
`

type UpdateAccountParams struct {
	Uuid     string
	Name     string
	Type     AccountType
	Metadata []byte
}

// UpdateAccount
//
//	   update accounts
//	      set name     = coalesce($2, name),
//	          type     = coalesce($3, type),
//	          metadata = coalesce($4, metadata)
//	    where uuid = $1
//	returning id, uuid, created_at, updated_at, name, type, metadata, ledger_id
func (q *Queries) UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error) {
	row := q.db.QueryRow(ctx, updateAccount,
		arg.Uuid,
		arg.Name,
		arg.Type,
		arg.Metadata,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Type,
		&i.Metadata,
		&i.LedgerID,
	)
	return i, err
}
