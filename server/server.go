package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ezhdanovskiy/docker-go-multi/env"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

type Server struct {
	db *sql.DB

	redis   *redis.Client
	channel string
	hash    string
}

func NewServer() (_ *Server, err error) {
	dataSourceName := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		env.PostgresUser, env.PostgresPassword, env.PostgresDatabase, env.PostgresHost, env.PostgresPort)
	log.Println("connect", dataSourceName)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("can't open db: %s", err)
	}

	log.Printf("db.Ping()")
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("bad ping: %s", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS values (number INT)")
	if err != nil {
		return nil, err
	}

	r := redis.NewClient(&redis.Options{
		Addr: env.RedisHost + ":" + env.RedisPort,
	})

	_, err = r.Ping().Result()
	if err != nil {
		return nil, err
	}

	log.Printf("Server created")
	return &Server{
		db:      db,
		redis:   r,
		channel: env.RedisChannel,
		hash:    env.RedisHash,
	}, nil
}

func (s *Server) Run(addr string) error {
	http.HandleFunc("/", s.index)
	http.HandleFunc("/values", s.values)
	http.HandleFunc("/values/current", s.valuesCurrent)
	http.HandleFunc("/values/all", s.valuesAll)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) values(w http.ResponseWriter, r *http.Request) {
	log.Println("method:", r.Method, r.URL)
	err := r.ParseForm()
	if err != nil {
		sendErr(err, w, "can't parse form", http.StatusBadRequest)
		return
	}
	log.Println("form:", r.Form)

	for _, index := range r.Form["index"] {
		ind, err := strconv.ParseInt(index, 10, 64)
		if err != nil {
			sendErr(err, w, "can't parse index", http.StatusBadRequest)
			return
		}
		if ind > 48 {
			sendErr(fmt.Errorf("index too high"), w, "(", http.StatusUnprocessableEntity)
			return
		}

		log.Printf("set %q to redis hash %q", index, s.channel)
		s.redis.HSet(s.hash, index, "Nothing yet!")

		log.Printf("publish %q to redis channel %q", index, s.channel)
		s.redis.Publish(s.channel, index)

		log.Printf("insert %q to db", index)
		_, err = s.db.Exec(`INSERT INTO values(number) VALUES($1)`, ind)
		if err != nil {
			sendErr(err, w, "db: can't insert index", http.StatusInternalServerError)
			return
		}
		break
	}
}

func (s *Server) valuesCurrent(w http.ResponseWriter, r *http.Request) {
	log.Println("method:", r.Method, r.URL)
	body, err := json.Marshal(s.redis.HGetAll(s.hash).Val())
	if err != nil {
		sendErr(err, w, "redis: can't marshal values", http.StatusInternalServerError)
		return
	}
	log.Printf("response: %q", body)
	fmt.Fprintf(w, "%s", body)
}

func (s *Server) valuesAll(w http.ResponseWriter, r *http.Request) {
	log.Println("method:", r.Method, r.URL)
	log.Printf("SELECT * FROM values;")
	rows, err := s.db.Query("SELECT * FROM values;")
	if err != nil {
		sendErr(err, w, "db: can't select values", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var indexes = make([]int64, 0)
	for rows.Next() {
		var index int64
		err := rows.Scan(&index)
		if err != nil {
			sendErr(err, w, "db: can't scan index", http.StatusInternalServerError)
			return
		}
		log.Println("index =", index)
		indexes = append(indexes, index)
	}
	body, err := json.Marshal(indexes)
	if err != nil {
		sendErr(err, w, "db: can't marshal indexes", http.StatusInternalServerError)
		return
	}
	log.Printf("response: %q", body)
	fmt.Fprintf(w, "%s", body)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	log.Println("method:", r.Method, r.URL)
	fmt.Fprintf(w, "Hi from server")
}

func (s *Server) Close() {
	err := s.db.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	err = s.redis.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func sendErr(err error, w http.ResponseWriter, msg string, statusCode int) {
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprintf(w, `{"error": "%s: %s"}`, msg, err)
		log.Printf(`{"error": "%s: %s"}`, msg, err)
	}
}
