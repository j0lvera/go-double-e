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

create trigger account_balance_updated_at
    before update
    on account_balances
    for each row
execute procedure set_updated_at();

-- Index for quick balance lookups
create index account_balances_lookup_idx on account_balances (account_id, created_at desc);

-- trigger function to create balance entries
create or replace function create_balance_entry()
    returns trigger as
$$
declare
    v_previous_balance bigint;
    v_new_balance      bigint;
    v_transaction_id   bigint;
    v_ledger_id        bigint;
    v_user_id          bigint;
    v_account_type     account_type;
begin
    -- get the transaction_id and ledger_id for denormalization
    select transaction_id into v_transaction_id from entries where id = NEW.id;
    select ledger_id, user_id into v_ledger_id, v_user_id from transactions where id = v_transaction_id;

    -- get account type
    select type into v_account_type from accounts where id = new.account_id;

    -- get previous balance
    select balance
    into v_previous_balance
    from account_balances
    where account_id = NEW.account_id
    order by created_at desc
    limit 1;

    -- calculate new balance based on account type and entry direction
    v_new_balance := v_previous_balance +
                     case
                         when v_account_type = 'asset' and NEW.direction = 'debit' then NEW.amount
                         when v_account_type = 'asset' and NEW.direction = 'credit' then -NEW.amount
                         when v_account_type = 'liability' and NEW.direction = 'debit' then -NEW.amount
                         when v_account_type = 'liability' and NEW.direction = 'credit' then NEW.amount
                         end;

    -- insert the balance entry
    insert into account_balances (balance,
                                  account_id,
                                  entry_id,
                                  transaction_id,
                                  ledger_id,
                                  user_id)
    values (v_new_balance,
            NEW.account_id,
            NEW.id,
            v_transaction_id,
            v_ledger_id,
            v_user_id);

    return NEW;
end;
$$ language plpgsql;

create trigger entry_create_balance
    after insert
    on entries
    for each row
execute function create_balance_entry();

-- helper function to get current balance
create or replace function get_account_balance(p_account_id bigint)
    returns bigint as
$$
select balance
from account_balances
where account_id = p_account_id
order by created_at desc
limit 1;
$$ language sql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop trigger entry_create_balance on entries;
drop function create_balance_entry();
drop function get_account_balance(bigint);

drop index account_balances_lookup_idx;
drop trigger account_balance_updated_at on account_balances;
drop table account_balances;
-- +goose StatementEnd
