package main

import (
	"context"
	"edtronautinterview/api/execution/executionhandler"
	"edtronautinterview/config"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

func main() {
	ctx := context.Background()

	c := config.New()

	fmtDbString := "host=%s user=%s password=%s dbname=%s port=%d sslmode=disable"
	dbString := fmt.Sprintf(fmtDbString, c.Postgres.Host, c.Postgres.User, c.Postgres.Password, c.Postgres.Name, c.Postgres.Port)

	pool, err := pgxpool.New(ctx, dbString)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %s", err.Error())
		return
	}
	defer pool.Close()

	s := asynq.NewServer(
		asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)},
		asynq.Config{
			Concurrency:    c.Redis.Concurrency,
			IsFailure:      func(err error) bool { return !executionhandler.IsRateLimitError(err) },
			RetryDelayFunc: executionhandler.RetryDelay,
		},
	)
	h := executionhandler.New(pool, c.Redis.ExecutionMemoryLimit, rate.NewLimiter(10, 30))

	if err := s.Run(h); err != nil {
		log.Fatal(err)
	}
}
