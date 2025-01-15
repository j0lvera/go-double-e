-- name: GetTransaction :one
select *
  from transactions
 where id = $1
 limit 1;

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
