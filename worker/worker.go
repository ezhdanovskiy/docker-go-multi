package worker

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis"
)

type Worker struct {
	r  *redis.Client
	ps *redis.PubSub
}

func NewWorker() (*Worker, error) {
	r := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})

	_, err := r.Ping().Result()
	if err != nil {
		return nil, err
	}

	pubsub := r.Subscribe("message")

	log.Println("Worker created")
	return &Worker{
		r:  r,
		ps: pubsub,
	}, nil
}

func (w Worker) Run() {
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

		w.r.HSet("values", msg.Payload, value)
	}
}

func (w *Worker) Close() {
	err := w.ps.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	err = w.r.Close()
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
