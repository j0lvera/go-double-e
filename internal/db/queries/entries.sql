-- name: GetEntry :one
select *
from entries
where id = $1
limit 1;

-- name: CreateEntry :one
insert into entries (amount, direction, transaction_id, account_id)
values ($1, $2, $3, $4)
returning *;

