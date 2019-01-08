package worker

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

type Worker struct {
	redis   *redis.Client
	pubsub  *redis.PubSub
	channel string
	hash    string
	wg      sync.WaitGroup
}

func NewWorker(addr, channel, hash string) (*Worker, error) {
	r := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	log.Printf("redis.Ping()")
	_, err := r.Ping().Result()
	for ; err != nil; _, err = r.Ping().Result() {
		log.Printf("bad redis ping: %s", err)
		time.Sleep(time.Second)
	}

	pubsub := r.Subscribe(channel)

	worker := &Worker{
		redis:   r,
		pubsub:  pubsub,
		channel: channel,
		hash:    hash,
	}

	go worker.Run()

	log.Println("Worker created")
	return worker, nil
}

func (w *Worker) Run() {
	for true {
		msg := <-w.pubsub.Channel()
		if msg == nil {
			log.Println("Stop listening messages")
			break
		}
		log.Printf("got %q from %q\n", msg.Payload, msg.Channel)
		w.wg.Add(1)
		go w.calcFib(msg.Payload)
	}
}

func (w *Worker) Close() {
	log.Println("Close redis pubsub")
	if err := w.pubsub.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	log.Println("Wait workers...")
	w.wg.Wait()

	log.Println("Close redis")
	if err := w.redis.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (w *Worker) calcFib(msg string) {
	defer w.wg.Done()
	index, err := strconv.ParseInt(msg, 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	log.Printf("start calculate fib(%d)\n", index)
	value := fib(index)
	log.Printf("finish fib(%d) = %d\n", index, value)

	log.Printf("put %q - %d to hash %q\n", msg, value, w.hash)
	w.redis.HSet(w.hash, msg, value)
}
