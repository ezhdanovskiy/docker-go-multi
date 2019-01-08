package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ezhdanovskiy/docker-go-multi/client"
	"github.com/ezhdanovskiy/docker-go-multi/env"
	"github.com/ezhdanovskiy/docker-go-multi/server"
	"github.com/ezhdanovskiy/docker-go-multi/worker"
)

func main() {
	var (
		httpAddr  = flag.String("http", ":8080", "Listen address")
		component = flag.String("component", "", "Run one component")
	)
	flag.Parse()

	run(*component, *httpAddr)
	log.Printf("Bye bye!")
}

func run(comp, addr string) {
	switch comp {
	case "client":
		cl, err := client.StartClient(addr)
		check(err)
		defer cl.Close()

	case "server":
		srv, err := server.StartServer(addr)
		check(err)
		defer srv.Close()

	case "worker":
		wkr, err := worker.NewWorker(env.RedisHost+":"+env.RedisPort, env.RedisChannel, env.RedisHash)
		check(err)
		defer wkr.Close()
	default:
		flag.Usage()
		os.Exit(2)
	}
	waitSignal()
}

func waitSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	log.Println("Received signal:", sig)
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
