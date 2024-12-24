-- +goose Up
-- +goose StatementBegin
create table users
(
    id         bigint generated always as identity primary key,
    uuid       text        not null default concat('u_', nanoid(10)),

    created_at timestamptz not null default current_timestamp,
    updated_at timestamptz not null default current_timestamp,
    email      text        not null check (char_length(email) < 255) unique,

    -- bcrypt hash is always 60
    password   text        not null check (char_length(password) = 60),

    -- create a unique index for the uuid
    unique (uuid)
);

create trigger user_updated_at
    before update
    on users
    for each row
execute procedure set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop trigger user_updated_at on users;
drop table users;
-- +goose StatementEnd
