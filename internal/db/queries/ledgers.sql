-- name: GetLedger :one
select *
  from ledgers
 where id = $1
 limit 1;

-- name: CreateLedger :one
   insert into ledgers (name, description, metadata)
   values ($1, $2, $3)
returning *;

-- name: ListLedgers :many
select *
  from ledgers
 where metadata @> $1::jsonb;


