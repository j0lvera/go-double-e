-- name: GetTransaction :one
select *
  from transactions
 where id = $1
 limit 1;

-- name: UpdateTransaction :one
     with credit_account as (select id from accounts where accounts.uuid = sqlc.narg('credit_account_uuid')::text),
          debit_account as (select id from accounts where accounts.uuid = sqlc.narg('debit_account_uuid')::text),
          ledger as (select id from ledgers where ledgers.uuid = sqlc.narg('ledger_uuid')::text)
   update transactions
      set amount            = coalesce(sqlc.narg('amount')::bigint, amount),
          date              = coalesce(sqlc.narg('date'), date),
          description       = coalesce(sqlc.narg('description'), description),
          metadata          = coalesce(sqlc.narg('metadata'), metadata),
          credit_account_id = coalesce((select id from credit_account), credit_account_id),
          debit_account_id  = coalesce((select id from debit_account), debit_account_id),
          ledger_id         = coalesce((select id from ledger), ledger_id)
    where transactions.uuid = sqlc.arg('uuid')
returning *;


-- name: CreateTransaction :one
     WITH credit_account AS (SELECT id
                               FROM accounts
                              WHERE accounts.uuid = sqlc.arg(credit_account_uuid)::text),
          debit_account AS (SELECT id
                              FROM accounts
                             WHERE accounts.uuid = sqlc.arg(debit_account_uuid)::text),
          ledger_id AS (SELECT id
                          FROM ledgers
                         WHERE ledgers.uuid = sqlc.arg(ledger_uuid)::text)
   INSERT
     INTO transactions (amount,
                        date,
                        description,
                        metadata,
                        credit_account_id,
                        debit_account_id,
                        ledger_id)
   VALUES (sqlc.arg(amount)::bigint,
           sqlc.arg(date)::date,
           sqlc.arg(description)::text,
           sqlc.arg(metadata)::jsonb,
           (SELECT id FROM credit_account),
           (SELECT id FROM debit_account),
           (SELECT id FROM ledger_id))
RETURNING *;

-- name: ListTransactions :many
  with ledger as (select id from ledgers where uuid = sqlc.arg(ledger_uuid)::text)
select uuid, amount, date, description, metadata
  from transactions
 where ledger_id = (select id from ledger)
   and metadata @> sqlc.arg(metadata)::jsonb;


-- name: DeleteTransaction :exec
delete
  from transactions
 where uuid = $1;
