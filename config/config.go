package config

import (
	"log"
	"time"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	Server   ConfServer
	Postgres ConfPostgres
	Redis    ConfRedis
}

type ConfServer struct {
	Port         int           `env:"SERVER_PORT,required"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
	Debug        bool          `env:"SERVER_DEBUG,required"`
}

type ConfPostgres struct {
	Host     string `env:"POSTGRES_HOST,required"`
	Port     int    `env:"POSTGRES_PORT,required"`
	Name     string `env:"POSTGRES_NAME,required"`
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	Debug    bool   `env:"POSTGRES_DEBUG,required"`
}

type ConfRedis struct {
	Host                 string        `env:"REDIS_HOST,required"`
	Port                 int           `env:"REDIS_PORT,required"`
	Concurrency          int           `env:"REDIS_CONCURRENCY,required"`
	ExecutionTimeout     time.Duration `env:"EXECUTION_TIMEOUT,required"`
	ExecutionMaxRetry    int           `env:"EXECUTION_MAX_RETRY,required"`
	ExecutionMemoryLimit string        `env:"EXECUTION_MEMORY_LIMIT,required"`
}

func New() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}

func NewPostgres() *ConfPostgres {
	var c ConfPostgres

	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
