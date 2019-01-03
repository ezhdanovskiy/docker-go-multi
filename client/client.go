package client

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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

func (c *Client) Run(addr string) error {
	http.HandleFunc("/", c.index)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) index(w http.ResponseWriter, r *http.Request) {
	log.Println("method:", r.Method)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			sendErr(err, w, "http: can't parse form")
		}

		for _, v := range r.Form["value"] {
			log.Printf("publish %s to %q\n", v, c.channel)
			c.redis.Publish(c.channel, v)
		}
	}

	data := struct {
		Indexes []string
		Values  map[string]string
	}{
		Values: getValues(),
	}

	for k := range data.Values {
		data.Indexes = append(data.Indexes, k)
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		sendErr(err, w, "http: can't execute template")
	}
}

func getValues() map[string]string {
	resp, err := http.Get("http://api:8080/values/current")
	if err != nil {
		log.Printf("error: can't get /api/values/current: %s", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error: can't read body : %s", err)
		return nil
	}
	log.Printf("body: %q", body)

	v := make(map[string]string)
	err = json.Unmarshal(body, &v)
	if err != nil {
		log.Printf("error: can't read body : %s", err)
		return nil
	}
	return v
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

func (c *Client) Close() {
	err := c.redis.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func sendErr(err error, w http.ResponseWriter, msg string) {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s: %s"}`, msg, err)
		log.Printf(`{"error": "%s: %s"}`, msg, err)
	}
}
