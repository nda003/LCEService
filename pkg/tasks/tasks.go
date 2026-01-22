package tasks

import (
	"edtronautinterview/pkg/db"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeCompilePython = "compile:python"
	TypeCompileGo     = "compile:go"
)

type CompilePayload struct {
	ExecutionId uuid.UUID   `json:"execution_id"`
	Language    db.Language `json:"language"`
	SourceCode  string      `json:"source_code"`
}

func NewCompileTask(executionId uuid.UUID, language db.Language, sourceCode string) (*asynq.Task, error) {
	payload, err := json.Marshal(CompilePayload{
		ExecutionId: executionId,
		Language:    language,
		SourceCode:  sourceCode,
	})
	if err != nil {
		return nil, err
	}

	switch language {
	case db.LanguagePython:
		return asynq.NewTask(TypeCompilePython, payload), nil
	case db.LanguageGo:
		return asynq.NewTask(TypeCompileGo, payload), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", string(language))
	}
}
