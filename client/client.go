package client

import (
	"fmt"
	"net/http"
)

type Client struct {
}

func NewClient() *Client {
	fmt.Println("Client created")
	return &Client{}
}

func (Client) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Hi from client")
}
