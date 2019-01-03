package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ezhdanovskiy/docker-go-multi/client"
	"github.com/ezhdanovskiy/docker-go-multi/env"
	"github.com/ezhdanovskiy/docker-go-multi/server"
	"github.com/ezhdanovskiy/docker-go-multi/worker"
)

var (
	httpAddr  = flag.String("http", ":8080", "Listen address")
	component = flag.String("component", "", "Run one component")
)

func main() {
	flag.Parse()

	switch *component {
	case "client":
		cl, err := client.NewClient(env.RedisHost+":"+env.RedisPort, env.RedisChannel, env.RedisHash)
		check(err)
		defer cl.Close()

		http.HandleFunc("/", cl.Index)
		err = http.ListenAndServe(*httpAddr, nil)
		check(err)

	case "server":
		srv, err := server.NewServer()
		check(err)
		defer srv.Close()
		err = srv.Run(*httpAddr)
		check(err)

	case "worker":
		wkr, err := worker.NewWorker(env.RedisHost+":"+env.RedisPort, env.RedisChannel, env.RedisHash)
		check(err)
		defer wkr.Close()
		wkr.Run()
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
