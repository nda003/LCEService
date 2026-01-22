-- +goose Up
create type code_session_status as enum ('active', 'inactive');
create type language as enum ('python', 'go');

create table if not exists code_sessions (
    id uuid default uuidv7() primary key,
    status code_session_status not null default 'active',
    language language not null default 'python',
    source_code text not null default ''
);

create type execution_status as enum ('queued', 'running', 'completed', 'failed', 'timeout');

create table if not exists executions (
    id uuid default uuidv7() primary key,
    status execution_status not null default 'queued',
    stdout text,
    stderr text,
    execution_time_ms smallint
);

create table if not exists execution_logs (
    id bigint generated always as identity primary key,
    execution_id uuid not null,
    status execution_status not null,
    timestamp timestamp without time zone default (now() at time zone 'utc'),
    constraint fk_execution_logs foreign key (execution_id) references executions (id) on delete cascade
);

-- +goose Down
drop table if exists code_sessions;
drop table if exists executions;
drop table if exists execution_logs;

drop type if exists code_session_status;
drop type if exists language;
drop type if exists execution_status;
