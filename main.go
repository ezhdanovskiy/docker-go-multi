package main

import (
	"flag"
	"log"
	"net/http"

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
		http.Handle("/client", client.NewClient())
		log.Fatal(http.ListenAndServe(*httpAddr, nil))
	case "server":
		http.Handle("/server", server.NewServer())
		log.Fatal(http.ListenAndServe(*httpAddr, nil))
	case "worker":
		worker.Worker{}.Run()
	default:
		go worker.Worker{}.Run()
		http.Handle("/client", client.NewClient())
		http.Handle("/server", server.NewServer())
		log.Fatal(http.ListenAndServe(*httpAddr, nil))
	}
}
