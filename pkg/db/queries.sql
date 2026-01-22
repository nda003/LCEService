-- name: CreateCodeSession :one
insert into code_sessions (language, source_code)
values ($1, $2)
returning *;

-- name: UpdateCodeSession :one
update code_sessions
set language = $2,
    source_code = $3
where id = $1
returning *;

-- name: GetCodeSession :one
select * from code_sessions
where id = $1;

-- name: CreateExecution :one
insert into executions (status)
values ($1)
returning id, status;

-- name: UpdateExecution :exec
update executions
set status = $2
where id = $1;

-- name: GetExecution :one
select * from executions
where id = $1;

-- name: CompleteExecution :exec
update executions
set status = $2,
    stdout = $3,
    stderr = $4,
    execution_time_ms = $5
where id = $1;
