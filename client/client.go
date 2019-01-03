package client

import (
	"html/template"
	"log"
	"net/http"

	"github.com/go-redis/redis"
)

type Client struct {
	redis   *redis.Client
	channel string
	hash    string
}

func NewClient(addr, channel, hash string) (*Client, error) {
	r := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := r.Ping().Result()
	if err != nil {
		return nil, err
	}

	log.Println("Client created")
	return &Client{
		redis:   r,
		channel: channel,
		hash:    hash,
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
			log.Printf("publish %s to %q\n", v, c.channel)
			c.redis.Publish(c.channel, v)
		}
	}

	data := struct {
		Indexes []string
		Values  map[string]string
	}{
		Values: make(map[string]string),
	}

	hashVal := c.redis.HGetAll(c.hash).Val()
	log.Printf("hashVal sise: %d\n", len(hashVal))
	for k, v := range hashVal {
		data.Indexes = append(data.Indexes, k)
		data.Values[k] = v
	}

	err := tmpl.Execute(w, data)
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
  <div>
    <form action="/" method="post">
      <label>Enter your value:</label>
      <input type="text" name="value">
      <input type="submit" value="Submit">
    </form>
    <h3>Indexes I have seen:</h3>
    {{range .Indexes}}
      <li>{{.}}</li>
    {{end}}
    <h3>Calculated values</h3>
    {{range $key, $value := .Values}}
      <li>{{$key}} - {{$value}}</li>
    {{end}}
  </div>
</body>
</html>
`))
