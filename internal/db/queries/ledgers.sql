-- name: GetLedger :one
select *
  from ledgers
 where uuid = $1
 limit 1;

-- name: CreateLedger :one
   insert into ledgers (name, description, metadata)
   values ($1, $2, $3)
returning *;


-- name: UpdateLedger :one
   update ledgers
      set name        = coalesce($2, name),
          description = coalesce($3, description),
          metadata    = coalesce($4, metadata)
    where uuid = $1
returning *;

-- name: ListLedgers :many
select *
  from ledgers
 where metadata @> $1::jsonb;


