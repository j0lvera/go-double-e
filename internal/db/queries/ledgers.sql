-- name: GetLedger :one
select *
from ledgers
where id = $1
limit 1;

-- name: CreateLedger :one
insert into ledgers (name, description, metadata, user_id)
values ($1, $2, $3, $4)
returning *;

