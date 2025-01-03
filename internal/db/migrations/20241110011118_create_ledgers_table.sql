-- +goose Up
-- +goose StatementBegin
create table ledgers
(
    id          bigint generated always as identity primary key,
    uuid        text        not null default nanoid(10),

    created_at  timestamptz not null default current_timestamp,
    updated_at  timestamptz not null default current_timestamp,

    name        text        not null,
    description text,
    metadata    jsonb,

    -- constraints
    constraint ledgers_uuid_unique unique (uuid),
    constraint ledgers_name_length_check check (char_length(name) < 255),
    constraint ledgers_description_length_check check (char_length(description) < 255)
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
