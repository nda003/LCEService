package executionapi

import (
	"edtronautinterview/pkg/db"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExecutionApi struct {
	Queries *db.Queries
}

func (api *ExecutionApi) GetExecution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Failed to parse UUID", http.StatusInternalServerError)
		return
	}

	execution, err := api.Queries.GetExecution(ctx, id)
	if err != nil {
		http.Error(w, "Failed to get execution", http.StatusInternalServerError)
		return
	}

	if execution.Status == db.ExecutionStatusCompleted {
		json, err := json.Marshal(map[string]any{
			"execution_id":      execution.ID.String(),
			"status":            strings.ToUpper(string(execution.Status)),
			"stdout":            execution.Stdout.String,
			"stderr":            execution.Stderr.String,
			"execution_time_ms": execution.ExecutionTimeMs.Int16,
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
	} else {
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
}

func New(pool *pgxpool.Pool) *ExecutionApi {
	return &ExecutionApi{Queries: db.New(pool)}
}
