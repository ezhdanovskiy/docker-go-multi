package env

import (
	"os"
)

var (
	RedisHost    = "redis"
	RedisPort    = "6379"
	RedisChannel = "message"
	RedisHash    = "values"
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
}
