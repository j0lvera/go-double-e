-- +goose Up
-- +goose StatementBegin
create type transaction_status as enum ('pending', 'posted');

create table transactions
(
    id                bigint generated always as identity primary key,
    uuid              text        not null default nanoid(10),

    created_at        timestamptz not null default current_timestamp,
    updated_at        timestamptz not null default current_timestamp,

    amount            bigint      not null default 0,
    date              date,
    description       text,
    metadata          jsonb,

    credit_account_id bigint      not null references accounts (id) on delete cascade,
    debit_account_id  bigint      not null references accounts (id) on delete cascade,

    -- links back to ledger and user for easy querying
    ledger_id         bigint      not null references ledgers (id) on delete cascade,

    -- constraints
    constraint transactions_uuid_unique unique (uuid),
    constraint transactions_amount_positive check (amount >= 0),
    constraint transactions_different_accounts check (credit_account_id != debit_account_id),
    constraint transactions_description_length_check check (char_length(description) < 255)
);

create trigger transaction_updated_at
    before update
    on transactions
    for each row
execute procedure set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop trigger transaction_updated_at on transactions;
drop table transactions;
drop type transaction_status;
-- +goose StatementEnd
