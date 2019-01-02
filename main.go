package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ezhdanovskiy/docker-go-multi/client"
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
		check(runClient())
		err := http.ListenAndServe(*httpAddr, nil)
		check(err)

	case "server":
		check(runServer())
		err := http.ListenAndServe(*httpAddr, nil)
		check(err)

	case "worker":
		check(runWorker())

	default:
		go func() {
			check(runWorker())
		}()
		check(runClient())
		check(runServer())
		err := http.ListenAndServe(*httpAddr, nil)
		check(err)
	}
}

func runClient() error {
	cl, err := client.NewClient()
	if err != nil {
		return err
	}
	http.HandleFunc("/", cl.Index)
	return nil
}

func runServer() error {
	srv, err := server.NewServer()
	if err != nil {
		return err
	}
	http.Handle("/server", srv)
	return nil
}

func runWorker() error {
	wkr, err := worker.NewWorker()
	if err != nil {
		return err
	}
	wkr.Run()
	return nil
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
