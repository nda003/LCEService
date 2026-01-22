-- +goose Up
-- +goose StatementBegin
create or replace function log_execution_status() returns trigger as $$
begin
    insert into execution_logs (execution_id, status)
    values (new.id, new.status);

    return new;
end;
$$ language plpgsql;
-- +goose StatementEnd

create trigger log_execution_insert_trigger
after insert or update on executions
for each row execute function log_execution_status();

-- +goose Down
drop function log_execution_status() cascade;
