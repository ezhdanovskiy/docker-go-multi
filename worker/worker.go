package worker

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis"
)

const (
	redisChannelDefault = "message"
	redisHashDefault    = "values"
)

type Worker struct {
	rd      *redis.Client
	ps      *redis.PubSub
	channel string
	hash    string
}

func NewWorker() (*Worker, error) {
	r := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})

	_, err := r.Ping().Result()
	if err != nil {
		return nil, err
	}

	channel := os.Getenv("REDIS_CHANNEL")
	if channel == "" {
		channel = redisChannelDefault
	}
	pubsub := r.Subscribe(channel)

	hash := os.Getenv("REDIS_HASH")
	if hash == "" {
		hash = redisHashDefault
	}

	log.Println("Worker created")
	return &Worker{
		rd:      r,
		ps:      pubsub,
		channel: channel,
		hash:    hash,
	}, nil
}

func (w *Worker) Run() {
	log.Println("Worker working...")
	for i := 0; i < 1000; i++ {
		msg, err := w.ps.ReceiveMessage()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		log.Printf("got %q from %q\n", msg.Payload, msg.Channel)
		index, err := strconv.ParseInt(msg.Payload, 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		log.Printf("start calculate fib(%d)\n", index)
		value := fib(index)
		log.Printf("finish fib(%d) = %d\n", index, value)

		log.Printf("put %q - %d to hash %q\n", msg.Payload, value, w.hash)
		w.rd.HSet(w.hash, msg.Payload, value)
	}
}

func (w *Worker) Close() {
	err := w.ps.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	err = w.rd.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func fib(index int64) int64 {
	if index < 2 {
		return 1
	}
	return fib(index-1) + fib(index-2)
}
