-- name: GetAccount :one
select *
  from accounts
 where id = $1
 limit 1;

-- name: CreateAccount :one
   insert into accounts (name, type, metadata, ledger_id)
   values ($1, $2, $3, $4)
returning *;

