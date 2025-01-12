# `go-double-e`

A simple double-entry web API written in Go.

## Notes

Query to test the `create_transactions_with_entries` postgres function:

```sql
    with ledger as (select ledgers.id as ledger_id from ledgers where ledgers.uuid = 'Oo44hlWDzz')
select
  from create_transaction_with_entries(
          'Milk',
          (select ledger_id from ledger)::bigint,
          ARRAY[
              '{"amount": 1000, "direction": "credit", "account_id": 7}'::jsonb,
              '{"amount": 1000, "direction": "debit", "account_id": 8}'::jsonb
              ],
          '{ "user_id": 25 }'::jsonb
       );
```