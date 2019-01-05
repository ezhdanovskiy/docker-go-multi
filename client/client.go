package client

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Client struct {
	http *http.Server
}

func StartClient(addr string) (*Client, error) {
	httpServer := &http.Server{Addr: addr}

	client := &Client{http: httpServer}

	http.HandleFunc("/", client.index)

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	log.Println("Client started")
	return client, nil
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
			u := "http://nginx/api/values"
			log.Printf("POST %q to %q", v, u)
			resp, err := http.PostForm(u, url.Values{"index": []string{v}})
			if err != nil {
				log.Printf("error: can't post values: %s", err)
				break
			}
			resp.Body.Close()
			break
		}
	}

	data := struct {
		Indexes []int64
		Values  map[string]string
	}{
		Indexes: getIndexes(),
		Values:  getCurrentValues(),
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		sendErr(err, w, "http: can't execute template")
	}
}

func getIndexes() []int64 {
	u := "http://nginx/api/values/all"
	log.Printf("GET %q", u)
	resp, err := http.Get(u)
	if err != nil {
		log.Printf("error: can't get all values: %s", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error: can't read body : %s", err)
		return nil
	}
	log.Printf("body: %q", body)

	v := make([]int64, 0)
	err = json.Unmarshal(body, &v)
	if err != nil {
		log.Printf("error: can't read body : %s", err)
		return nil
	}
	return v
}

func getCurrentValues() map[string]string {
	u := "http://nginx/api/values/current"
	log.Printf("GET %q", u)
	resp, err := http.Get(u)
	log.Printf("GET /api/values/current")
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
	log.Println("Shutdown http")
	if err := c.http.Shutdown(nil); err != nil {
		fmt.Fprintln(os.Stderr, "http.Shutdown()", err)
	}
}

func sendErr(err error, w http.ResponseWriter, msg string) {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s: %s"}`, msg, err)
		log.Printf(`{"error": "%s: %s"}`, msg, err)
	}
}
