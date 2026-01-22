# Live Code Execution Service

A backend service implemented in Go for live code execution using Redis for queue-based worker management and PostgreSQL for data and log storage.

## Setup

This repository uses Docker compose for orchestration, thus requires Docker to be installed and configured. To build and start the service, run:

```bash
docker compose up --build
```

The service is exposed at `localhost:8080`.

## Design

### Architecture Overview

![Architecture diagram](./images/architecture.svg)
<img src="./images/architecture.svg">

The Go web server is the API to the Postgres server for creating and updating code sessions, as well as enqueuing an execution task to the Redis server's message queue. The Go consumer can then poll the tasks in the Redis message for execution, logging the status of the execution into the Postgres server as it goes. Interactions with Redis's task queue and task-worker concurrency are done with the Go library [Asynq](https://github.com/hibiken/asynq). Below is a sequence diagram of the service:

![Sequence diagram of the service](./images/sequence-diagram.svg)
<img src="./images/sequence-diagram.svg">

### Reliability & Data Model

An execution task's lifecycle begin when the Go web server call the Redis server to enqueue the task, logging the task into the Postgres server as `QUEUED`. When the task is polled by the Go consumer and executed, the Go consumer logs the execution as `RUNNING`. If the execution is running for more than 10 seconds (configurable in `.env`), the consumer logs the execution as `TIMEOUT`, these are then archived by Asynq into a dead letter queue. This is to prevent infinite loops from taking up a worker's time and resources. If the task failed due to an internal consumer error, the task is retried up to 5 times until the task is done successfully, or the task failed all 5 retries, logging the execution as `FAILED`. If the task is done successfully, the consumer logs the execution as `COMPLETED`, updating the execution in Postgres server with the execution's outputs and the execution time in milliseconds. Each execution lifecycle is logged with a timestamp using a log trigger in the Postgres server.

Because of timing constraint, executions are not currently being done in isolated environments. Unsafe and malicious code are able to affect the Go consumer. In the future, these limitations can be addressed by providing each tasks with a containerized Docker environment using gVisor for a more secured isolation.

Programming languages available are Python and Go.

### Scalability Considerations

A Go consumer can currently handle 4 execution tasks concurrently and can be configured for more concurrent workers as needed. This architecture can be horizontally scaled by adding more Redis and consumer clusters. Each execution task can be enqueued into a Redis node and polled by an available consumer node. This architecture ensures that if a consumer node fails or is busy, a task can be polled by a different consumer node, ensuring that a queued task is processed as quickly as possible. Redis clusters for separate tasks prevent backlog build up and redundancy in case of failure.

A potential bottleneck is the constant write operation done on the Postgres server, for example patching code sessions, updating and logging execution task. A better database for this would be a NoSQL database optimized for write operation such Apache Cassandra.

### Trade-offs

The architecture was designed with simplicity in mind for the purpose of quick prototyping and iterating. Postgres was chosen for its ecosystem and good documentation, and so was Redis. Go was selected as the programming language for its simplicity without sacrificing performance.

Production-readiness lacks a foundation for horizontal scaling such as a sophisticated orchestration and scaling configuration. Executions are not run in an isolated environment and thus vulnerable to malicious actors and adversaries. The service currently only supports Python and Go.

## API documentation

### POST `/code-sessions`

Create a new live coding session with an empty source code and the default programming language of Python.

#### Response

```json
{
  "session_id": "uuid",
  "status": "ACTIVE"
}
```

### PATCH `/code-sessions/{session_id}`

Update and save the code session's programming language and source code.

#### Request

```json
{
  "language": "python",
  "source_code": "print(\"Hello World\")"
}
```

#### Response

```json
{
  "session_id": "uuid",
  "status": "ACTIVE"
}
```

### POST `/code-sessions/{session_id}/run`

Execute the current code asynchronously.

#### Response

```json
{
  "execution_id": "uuid",
  "status": "QUEUED"
}
```

#### GET `/executions/{execution_id}`

Retrieve execution status and result.

#### Response when `QUEUED`, `RUNNING`, `FAILED`, or `TIMEOUT`

```json
{
  "execution_id": "uuid",
  "status": "RUNNING"
}
```

#### Response when `COMPLETED`

```json
{
  "execution_id": "uuid",
  "status": "COMPLETED",
  "stderr": "",
  "stdout": "Hello world\n",
  "execution_time_ms": 54
}
```

## Improvements

With more time, I would implement the containerized Docker environment system with gVisor for isolated and safe code execution. I would also configure Redis and Postgres for prototyping future horizontal scaling and orchestration, and maybe switch from Postgres to Cassandra for faster and more effiecent logging.
