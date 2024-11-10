-- +goose Up
-- +goose StatementBegin
create table account_balances
(
    id             bigint generated always as identity primary key,
    uuid           text        not null default concat('b_', nanoid(10)),

    created_at     timestamptz not null default current_timestamp,
    updated_at     timestamptz not null default current_timestamp,

    -- this is the balance after applying the entry
    balance        bigint      not null default 0,

    -- reference both the account and the entry that caused this balance change
    account_id     bigint      not null references accounts (id) on delete cascade,
    entry_id       bigint      not null references entries (id) on delete cascade,

    -- denormalized references for easier querying
    transaction_id bigint      not null references transactions (id) on delete cascade,
    ledger_id      bigint      not null references ledgers (id) on delete cascade,
    user_id        bigint      not null references users (id) on delete cascade,

    unique (uuid)
);

-- Index for quick balance lookups
create index account_balances_lookup_idx on account_balances (account_id, created_at desc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index account_balances_lookup_idx;
drop table account_balances;
-- +goose StatementEnd
