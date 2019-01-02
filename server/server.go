package server

import (
	"fmt"
	"net/http"
)

type Server struct {
}

func NewServer() *Server {
	fmt.Println("Server created")
	return &Server{}
}

func (Server) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Hi from server")
}
