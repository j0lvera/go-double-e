-- name: GetTransaction :one
select *
  from transactions
 where id = $1
 limit 1;

-- name: CreateTransaction :one
   insert into transactions (description, metadata, ledger_id)
   values ($1::text, $2::jsonb, $3::bigint)
returning *;


-- name: CreateTransactionWithEntries :one
  with ledger as (select ledgers.id as ledger_id from ledgers where ledgers.uuid = sqlc.arg(ledger_uuid)::text)
select t
  from create_transaction_with_entries(
               sqlc.arg(description)::text,
               (select ledger_id from ledger)::bigint,
               sqlc.arg(entries)::jsonb[],
               sqlc.arg(metadata)::jsonb
       ) as t;
-- select t
--   from create_transaction_with_entries(
--                sqlc.arg(description)::text,
--                (select ledger_id from ledger)::bigint,
--                sqlc.arg(entries)::jsonb[],
--                sqlc.arg(metadata)::jsonb
--        ) as t;

-- -- name: CreateAccount :one
--      with ledger as (select id
--                        from ledgers
--                       where uuid = sqlc.arg(ledger_uuid)::text)
--    insert
--      into accounts (name, type, metadata, ledger_id)
--    values ($1, $2, $3, (select id from ledger))
-- returning *;
