-- +goose Up
-- +goose StatementBegin
create table ledgers
(
    id          bigint generated always as identity primary key,
    uuid        text        not null default concat('l_', nanoid(10)),

    created_at  timestamptz not null default current_timestamp,
    updated_at  timestamptz not null default current_timestamp,

    name        text        not null check (char_length(name) < 255),
    description text,
    metadata    jsonb,

    user_id     bigint      not null references users (id) on delete cascade,

    -- create a unique index for the uuid
    unique (uuid)
);

create trigger ledger_updated_at
    before update
    on ledgers
    for each row
execute procedure set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop trigger ledger_updated_at on ledgers;
drop table ledgers;
-- +goose StatementEnd
