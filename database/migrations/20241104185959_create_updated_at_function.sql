-- +goose Up
-- +goose StatementBegin
create or replace function set_updated_at()
    returns trigger as
$$
begin
    new.updated_at := current_timestamp;
    return new;
end;
$$ language plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop function set_updated_at();
-- +goose StatementEnd
