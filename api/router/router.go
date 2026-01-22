package router

import (
	"edtronautinterview/api/codesessionapi"
	"edtronautinterview/api/execution/executionapi"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool, redisClient *asynq.Client, taskTimeout time.Duration, maxRetry int) *chi.Mux {
	r := chi.NewRouter()

	csapi := codesessionapi.New(pool, redisClient, taskTimeout, maxRetry)
	r.Post("/code-sessions", csapi.CreateCodeSession)
	r.Patch("/code-sessions/{id}", csapi.PatchCodeSession)
	r.Post("/code-sessions/{id}/run", csapi.RunCodeSession)

	exapi := executionapi.New(pool)
	r.Get("/executions/{id}", exapi.GetExecution)

	return r
}
