-- +goose Up
-- +goose StatementBegin
create type entry_direction as enum ('debit', 'credit');

create table entries
(
    id             bigint generated always as identity primary key,
    uuid           text            not null default nanoid(10),

    created_at     timestamptz     not null default current_timestamp,
    updated_at     timestamptz     not null default current_timestamp,

    amount         bigint          not null default 0,
    direction      entry_direction not null,

    transaction_id bigint          not null references transactions (id) on delete cascade,
    account_id     bigint          not null references accounts (id) on delete cascade,

    -- constraints
    constraint entries_uuid_unique unique (uuid),
    constraint entries_amount_positive check (amount >= 0)
);

create trigger entry_updated_at
    before update
    on entries
    for each row
execute procedure set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop trigger entry_updated_at on entries;
drop table entries;
drop type entry_direction;
-- +goose StatementEnd
