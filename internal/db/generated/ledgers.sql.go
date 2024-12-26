// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: ledgers.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLedger = `-- name: CreateLedger :one
insert into ledgers (name, description, metadata, user_id)
values ($1, $2, $3, $4)
returning id, uuid, created_at, updated_at, name, description, metadata, user_id
`

type CreateLedgerParams struct {
	Name        string
	Description pgtype.Text
	Metadata    []byte
	UserID      int64
}

// CreateLedger
//
//	insert into ledgers (name, description, metadata, user_id)
//	values ($1, $2, $3, $4)
//	returning id, uuid, created_at, updated_at, name, description, metadata, user_id
func (q *Queries) CreateLedger(ctx context.Context, arg CreateLedgerParams) (Ledger, error) {
	row := q.db.QueryRow(ctx, createLedger,
		arg.Name,
		arg.Description,
		arg.Metadata,
		arg.UserID,
	)
	var i Ledger
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Description,
		&i.Metadata,
		&i.UserID,
	)
	return i, err
}

const getLedger = `-- name: GetLedger :one
select id, uuid, created_at, updated_at, name, description, metadata, user_id
from ledgers
where id = $1
limit 1
`

// GetLedger
//
//	select id, uuid, created_at, updated_at, name, description, metadata, user_id
//	from ledgers
//	where id = $1
//	limit 1
func (q *Queries) GetLedger(ctx context.Context, id int64) (Ledger, error) {
	row := q.db.QueryRow(ctx, getLedger, id)
	var i Ledger
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Description,
		&i.Metadata,
		&i.UserID,
	)
	return i, err
}
