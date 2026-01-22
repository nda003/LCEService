package main

import (
	"context"
	"edtronautinterview/api/router"
	"edtronautinterview/config"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
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

	redisClient := asynq.NewClient(asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Fatalf("Failed to close Redis client: %s", err.Error())
		}
	}()

	r := router.New(pool, redisClient, c.Redis.ExecutionTimeout, c.Redis.ExecutionMaxRetry)
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Server.Port),
		Handler:      r,
		ReadTimeout:  c.Server.TimeoutRead,
		WriteTimeout: c.Server.TimeoutWrite,
		IdleTimeout:  c.Server.TimeoutIdle,
	}

	s.ListenAndServe()
}

func hello(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Hello world")
}
