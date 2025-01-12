-- name: GetAccount :one
select *
  from accounts
 where uuid = $1
 limit 1;

-- name: UpdateAccount :one
   update accounts
      set name     = coalesce($2, name),
          type     = coalesce($3, type),
          metadata = coalesce($4, metadata)
    where uuid = $1
returning *;


-- name: CreateAccount :one
     with ledger as (select id
                       from ledgers
                      where uuid = sqlc.arg(ledger_uuid)::text)
   insert
     into accounts (name, type, metadata, ledger_id)
   values ($1, $2, $3, (select id from ledger))
returning *;

-- name: ListAccounts :many
  with ledger as (select id from ledgers where uuid = sqlc.arg(ledger_uuid)::text)
select uuid, name, type, metadata
  from accounts
 where ledger_id = (select id from ledger)
   and metadata @> sqlc.arg(metadata)::jsonb;