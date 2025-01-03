-- name: GetTransaction :one
select *
  from transactions
 where id = $1
 limit 1;

-- name: CreateTransaction :one
   insert into transactions (description, metadata, ledger_id)
   values ($1, $2, $3)
returning *;

