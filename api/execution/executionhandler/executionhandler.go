package executionhandler

import (
	"bytes"
	"context"
	"edtronautinterview/pkg/db"
	"edtronautinterview/pkg/tasks"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

type ExecutionHandler struct {
	Queries     *db.Queries
	MemoryLimit string
	RateLimiter *rate.Limiter
}

type RateLimitError struct {
	RetryIn time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited (retry in  %v)", e.RetryIn)
}

func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

func RetryDelay(n int, err error, task *asynq.Task) time.Duration {
	var ratelimitErr *RateLimitError
	if errors.As(err, &ratelimitErr) {
		return ratelimitErr.RetryIn
	}
	return asynq.DefaultRetryDelayFunc(n, err, task)
}

func New(pool *pgxpool.Pool, memoryLimit string, limiter *rate.Limiter) *ExecutionHandler {
	return &ExecutionHandler{Queries: db.New(pool), MemoryLimit: memoryLimit, RateLimiter: limiter}
}

func (h *ExecutionHandler) failExecution(ctx context.Context, executionId uuid.UUID) error {
	if err := h.Queries.UpdateExecution(ctx, db.UpdateExecutionParams{ID: executionId, Status: db.ExecutionStatusFailed}); err != nil {
		return err
	}

	return nil
}

func (h *ExecutionHandler) compilePython(ctx context.Context, payload *tasks.CompilePayload) error {
	if err := h.Queries.UpdateExecution(ctx, db.UpdateExecutionParams{ID: payload.ExecutionId, Status: db.ExecutionStatusRunning}); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "python3", "-c", payload.SourceCode)

	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	start := time.Now()
	cmd.Run()
	execTime := time.Since(start)

	if err := h.Queries.CompleteExecution(ctx, db.CompleteExecutionParams{
		ID:              payload.ExecutionId,
		Status:          db.ExecutionStatusCompleted,
		Stdout:          pgtype.Text{String: outBuffer.String(), Valid: true},
		Stderr:          pgtype.Text{String: errBuffer.String(), Valid: true},
		ExecutionTimeMs: pgtype.Int2{Int16: int16(execTime.Milliseconds()), Valid: true},
	}); err != nil {
		return err
	}

	return nil
}

func (h *ExecutionHandler) compileGo(ctx context.Context, payload *tasks.CompilePayload) error {
	if err := h.Queries.UpdateExecution(ctx, db.UpdateExecutionParams{ID: payload.ExecutionId, Status: db.ExecutionStatusRunning}); err != nil {
		return err
	}

	dir := filepath.Join(os.TempDir(), payload.ExecutionId.String())

	err := os.Mkdir(dir, 0644)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	file, err := os.Create(filepath.Join(dir, "main.go"))
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	file.WriteString(payload.SourceCode)

	cmd := exec.CommandContext(ctx, "go", "run", file.Name())

	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	start := time.Now()
	cmd.Run()
	execTime := time.Since(start)

	if err := h.Queries.CompleteExecution(ctx, db.CompleteExecutionParams{
		ID:              payload.ExecutionId,
		Status:          db.ExecutionStatusCompleted,
		Stdout:          pgtype.Text{String: outBuffer.String(), Valid: true},
		Stderr:          pgtype.Text{String: errBuffer.String(), Valid: true},
		ExecutionTimeMs: pgtype.Int2{Int16: int16(execTime.Milliseconds()), Valid: true},
	}); err != nil {
		return err
	}

	return nil
}

func (h *ExecutionHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	if !h.RateLimiter.Allow() {
		return &RateLimitError{RetryIn: time.Duration(rand.Intn(10)) * time.Second}
	}

	var payload tasks.CompilePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.failExecution(ctx, payload.ExecutionId)
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := make(chan error, 1)

	switch t.Type() {
	case tasks.TypeCompilePython:
		go func() {
			select {
			case c <- h.compilePython(ctx, &payload):
			case <-ctx.Done():
				return
			}
		}()
	case tasks.TypeCompileGo:
		go func() {
			select {
			case c <- h.compileGo(ctx, &payload):
			case <-ctx.Done():
				return
			}
		}()
	default:
		return fmt.Errorf("unexpected task type: %s", t.Type())
	}

	select {
	case <-ctx.Done():
		if err := h.Queries.UpdateExecution(context.Background(), db.UpdateExecutionParams{
			ID:     payload.ExecutionId,
			Status: db.ExecutionStatusTimeout,
		}); err != nil {
			return err
		}

		return asynq.SkipRetry
	case err := <-c:
		if err != nil {
			log.Printf("%s", err.Error())
			h.failExecution(ctx, payload.ExecutionId)
		}

		return err
	}
}
