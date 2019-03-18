package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ezhdanovskiy/docker-go-multi/env"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq" // pg need only for server
)

// Server struct
type Server struct {
	db *sql.DB

	redis   *redis.Client
	channel string
	hash    string

	http *http.Server
}

// StartServer start server
func StartServer(addr string) (*Server, error) {
	dataSourceName := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		env.PostgresUser, env.PostgresPassword, env.PostgresDatabase, env.PostgresHost, env.PostgresPort)
	log.Println("connect", dataSourceName)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("can't open db: %s", err)
	}

	log.Printf("postgres.Ping()")
	err = db.Ping()
	for ; err != nil; err = db.Ping() {
		log.Printf("bad postgres ping: %s", err)
		time.Sleep(time.Second)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS values (number INT)")
	if err != nil {
		return nil, err
	}

	r := redis.NewClient(&redis.Options{
		Addr: env.RedisHost + ":" + env.RedisPort,
	})

	log.Printf("redis.Ping()")
	_, err = r.Ping().Result()
	for ; err != nil; _, err = r.Ping().Result() {
		log.Printf("bad redis ping: %s", err)
		time.Sleep(time.Second)
	}

	httpServer := &http.Server{Addr: addr}

	srv := &Server{
		db:      db,
		redis:   r,
		channel: env.RedisChannel,
		hash:    env.RedisHash,
		http:    httpServer,
	}

	http.HandleFunc("/", srv.index)
	http.HandleFunc("/values", srv.values)
	http.HandleFunc("/values/current", srv.valuesCurrent)
	http.HandleFunc("/values/all", srv.valuesAll)

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	log.Println("Server started")
	return srv, nil
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

// Close postgres, redis and http
func (s *Server) Close() {
	log.Println("Close postgres")
	if err := s.db.Close(); err != nil {
		log.Fatalln("db.Close():", err)
	}

	log.Println("Close redis")
	if err := s.redis.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "redis.Close()", err)
	}

	log.Println("Shutdown http")
	if err := s.http.Shutdown(nil); err != nil {
		fmt.Fprintln(os.Stderr, "http.Shutdown()", err)
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
