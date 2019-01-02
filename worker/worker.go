package worker

import (
	"fmt"
	"time"
)

type Worker struct{}

func (w Worker) Run() {
	fmt.Println("Worker created")
	for i := 0; i < 1000; i++ {
		fmt.Printf("%d ", i)
		time.Sleep(time.Second * 10)
	}
}
