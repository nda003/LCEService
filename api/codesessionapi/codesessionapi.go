package codesessionapi

import (
	"edtronautinterview/pkg/db"
	"edtronautinterview/pkg/tasks"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CodeSessionApi struct {
	Queries     *db.Queries
	Pool        *pgxpool.Pool
	RedisClient *asynq.Client
	TaskTimeout time.Duration
	MaxRetry    int
}

func (api *CodeSessionApi) CreateCodeSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	session, err := api.Queries.CreateCodeSession(ctx, db.CreateCodeSessionParams{
		Language:   "python",
		SourceCode: "",
	})
	if err != nil {
		http.Error(w, "Failed to create code session", http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(map[string]string{
		"session_id": session.ID.String(),
		"status":     strings.ToUpper(string(session.Status)),
	})

	if err != nil {
		http.Error(w, "Failed to encode json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(json); err != nil {
		http.Error(w, "Failed to write json", http.StatusInternalServerError)
	}
}

type PatchCodeSessionRequestPayload struct {
	Language   string `json:"language"`
	SourceCode string `json:"source_code"`
}

func (api *CodeSessionApi) PatchCodeSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Failed to parse UUID", http.StatusInternalServerError)
		return
	}

	var payload PatchCodeSessionRequestPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to parse payload", http.StatusInternalServerError)
		return
	}

	session, err := api.Queries.UpdateCodeSession(ctx, db.UpdateCodeSessionParams{
		ID:         id,
		Language:   db.Language(payload.Language),
		SourceCode: payload.SourceCode,
	})
	if err != nil {
		http.Error(w, "Failed to patch code session", http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(map[string]string{
		"session_id": session.ID.String(),
		"status":     strings.ToUpper(string(session.Status)),
	})
	if err != nil {
		http.Error(w, "Failed to encode json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(json); err != nil {
		http.Error(w, "Failed to write json", http.StatusInternalServerError)
	}
}

func (api *CodeSessionApi) RunCodeSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Failed to parse UUID", http.StatusInternalServerError)
		return
	}

	session, err := api.Queries.GetCodeSession(ctx, id)
	if err != nil {
		http.Error(w, "Failed to get code session", http.StatusInternalServerError)
		return
	}

	tx, err := api.Pool.Begin(ctx)
	if err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	qtx := api.Queries.WithTx(tx)

	execution, err := qtx.CreateExecution(ctx, db.ExecutionStatusQueued)
	if err != nil {
		http.Error(w, "Failed to create execution", http.StatusInternalServerError)
		return
	}

	task, err := tasks.NewCompileTask(execution.ID, session.Language, session.SourceCode)
	if err != nil {
		http.Error(w, "Failed to create code session running task", http.StatusInternalServerError)
		return
	}

	if _, err := api.RedisClient.EnqueueContext(
		ctx, task,
		asynq.Timeout(api.TaskTimeout),
		asynq.MaxRetry(api.MaxRetry),
	); err != nil {
		http.Error(w, "Failed to enqueue code session running task", http.StatusInternalServerError)
		return
	}

	tx.Commit(ctx)

	json, err := json.Marshal(map[string]string{
		"execution_id": execution.ID.String(),
		"status":       strings.ToUpper(string(execution.Status)),
	})
	if err != nil {
		http.Error(w, "Failed to encode json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(json); err != nil {
		http.Error(w, "Failed to write json", http.StatusInternalServerError)
	}
}

func New(pool *pgxpool.Pool, redisClient *asynq.Client, taskTimeout time.Duration, maxRetry int) *CodeSessionApi {
	return &CodeSessionApi{
		Queries:     db.New(pool),
		Pool:        pool,
		RedisClient: redisClient,
		TaskTimeout: taskTimeout,
		MaxRetry:    maxRetry,
	}
}
