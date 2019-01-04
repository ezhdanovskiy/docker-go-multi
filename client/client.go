package client

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
}

func NewClient() (*Client, error) {

	log.Println("Client created")
	return &Client{}, nil
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
	log.Println("method:", r.Method, r.URL)
	if r.URL.String() == "/favicon.ico" {
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			sendErr(err, w, "http: can't parse form")
		}
		log.Println("form:", r.Form)

		for _, v := range r.Form["value"] {
			log.Printf("post %q to /api/values", v)
			resp, err := http.PostForm("http://api:8080/values", url.Values{"index": []string{v}})
			if err != nil {
				log.Printf("error: can't post values: %s", err)
			}
			resp.Body.Close()

			break
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
	log.Printf("GET /api/values/current")
	resp, err := http.Get("http://api:8080/values/current")
	if err != nil {
		log.Printf("error: can't get current values: %s", err)
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
}

func sendErr(err error, w http.ResponseWriter, msg string) {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s: %s"}`, msg, err)
		log.Printf(`{"error": "%s: %s"}`, msg, err)
	}
}
