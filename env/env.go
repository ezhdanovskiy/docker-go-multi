package env

import (
	"os"
)

// Envs for all components
var (
	RedisHost    = "redis"
	RedisPort    = "6379"
	RedisChannel = "message"
	RedisHash    = "values"

	PostgresHost     = "postgres"
	PostgresPort     = "5432"
	PostgresUser     = "multi"
	PostgresPassword = "multi"
	PostgresDatabase = "multi"
)

func init() {
	if v := os.Getenv("REDIS_HOST"); v != "" {
		RedisHost = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		RedisPort = v
	}
	if v := os.Getenv("REDIS_CHANNEL"); v != "" {
		RedisChannel = v
	}
	if v := os.Getenv("REDIS_HASH"); v != "" {
		RedisHash = v
	}

	if v := os.Getenv("POSTGRES_HOST"); v != "" {
		PostgresHost = v
	}
	if v := os.Getenv("POSTGRES_PORT"); v != "" {
		PostgresPort = v
	}
	if v := os.Getenv("POSTGRES_USER"); v != "" {
		PostgresUser = v
	}
	if v := os.Getenv("POSTGRES_PASSWORD"); v != "" {
		PostgresPassword = v
	}
	if v := os.Getenv("POSTGRES_DB"); v != "" {
		PostgresDatabase = v
	}
}
