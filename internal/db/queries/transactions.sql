-- name: GetTransaction :one
select *
from transactions
where id = $1
limit 1;

-- name: CreateTransaction :one
insert into transactions (description, metadata, ledger_id, user_id)
values ($1, $2, $3, $4)
returning *;

