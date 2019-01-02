package client

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis"
)

const messageChannel = "message"

type Client struct {
	r *redis.Client
}

func NewClient() (*Client, error) {
	r := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})

	_, err := r.Ping().Result()
	if err != nil {
		return nil, err
	}

	log.Println("Client created")
	return &Client{
		r: r,
	}, nil
}

func (c *Client) Index(w http.ResponseWriter, r *http.Request) {
	log.Println("method:", r.Method) //get request method
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Printf("error: can't parce form %s\n", err)
		}
		// logic part of log in
		for _, v := range r.Form["value"] {
			log.Printf("publish %s to %q\n", v, messageChannel)
			c.r.Publish(messageChannel, v)
		}
	}
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("error: can't execute template: %s\n", err)
	}
}

var tmpl = template.Must(template.New("tmpl").Parse(`
<html>
<head>
    <title>Fib calculator</title>
</head>
<body>
<form action="/" method="post">
    Enter your value:<input type="text" name="value">
    <input type="submit" value="Submit">
</form>
</body>
</html>
`))
