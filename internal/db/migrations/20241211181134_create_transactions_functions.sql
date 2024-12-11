-- +goose Up
-- +goose StatementBegin
create or replace function create_transaction_with_entries(
    p_description text,
    p_ledger_id bigint,
    p_user_id bigint,
    p_entries json[],
    p_metadata jsonb default null
) returns bigint as
$$
declare
    v_transaction_id bigint;
    v_entry          json;
    v_total_balance  bigint := 0;
begin
    -- calculate the total balance of the entries
    foreach v_entry in array p_entries
        loop
        -- for each entry pair:
        -- debit (asset): +amount
        -- credit (asset): -amount
        -- debit (liability): -amount
        -- credit (liability): +amount
            select v_total_balance + (
                case
                    when (select type from accounts where id = (v_entry ->> 'account_id')::bigint) = 'asset'
                        then (v_entry ->> 'amount')::bigint
                    else -(v_entry ->> 'amount')::bigint
                    end
                ) - (
                       case
                           when (select type from accounts where id = (v_entry ->> 'account_id')::bigint) =
                                'asset'
                               then (v_entry ->> 'amount')::bigint
                           else -(v_entry ->> 'amount')::bigint
                           end
                       )
            into v_total_balance;
        end loop;

    if v_total_balance != 0 then
        raise exception 'Total balance of entries must be 0';
    end if;

    -- create transaction and entries if balanced
    insert into transactions (description, ledger_id, user_id, metadata)
    values (p_description, p_ledger_id, p_user_id, p_metadata)
    returning id into v_transaction_id;

    foreach v_entry in array p_entries
        loop
            insert into entries (transaction_id, account_id, amount, direction)
            values (v_transaction_id, (v_entry ->> 'account_id')::bigint, (v_entry ->> 'amount')::bigint,
                    'debit'),
                   (v_transaction_id, (v_entry ->> 'account_id')::bigint, (v_entry ->> 'amount')::bigint,
                    'credit');
        end loop;

    return v_transaction_id;
end;
$$ language plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop function create_transaction_with_entries(text, bigint, bigint, json[], jsonb);
-- +goose StatementEnd
