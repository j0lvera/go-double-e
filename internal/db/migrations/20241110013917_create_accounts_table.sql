-- +goose Up
-- +goose StatementBegin
create type account_type as enum ('asset', 'liability');

create table accounts
(
    id         bigint generated always as identity primary key,
    uuid       text         not null default concat('a_', nanoid(10)),

    created_at timestamptz  not null default current_timestamp,
    updated_at timestamptz  not null default current_timestamp,

    name       text         not null check (char_length(name) < 255),
    type       account_type not null,
    balance    bigint       not null default 0,
    metadata   jsonb,

    ledger_id  bigint       not null references ledgers (id) on delete cascade,
    user_id    bigint       not null references users (id) on delete cascade,

    unique (uuid)
);

create trigger ledger_updated_at
    before update
    on accounts
    for each row
execute procedure set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop trigger ledger_updated_at on accounts;
drop table accounts;
-- +goose StatementEnd
