-- +goose Up
-- +goose StatementBegin
create type account_type as enum ('asset', 'liability', 'equity', 'revenue', 'expense');

create table accounts
(
    id         bigint generated always as identity primary key,
    uuid       text         not null default nanoid(10),

    created_at timestamptz  not null default current_timestamp,
    updated_at timestamptz  not null default current_timestamp,

    name       text         not null,
    type       account_type not null,
    metadata   jsonb,

    ledger_id  bigint       not null references ledgers (id) on delete cascade,

    -- constraints
    constraint accounts_uuid_unique unique (uuid),
    constraint accounts_name_length_check check (char_length(name) < 255)
);

create trigger account_updated_at
    before update
    on accounts
    for each row
execute procedure set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop trigger account_updated_at on accounts;
drop table accounts;
drop type account_type;
-- +goose StatementEnd
