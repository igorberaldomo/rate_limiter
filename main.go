package main

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", perClientLimiter(handler))
	r.Get("/{API_KEY}", perTokenLimiter(handler))

	http.ListenAndServe(":8080", nil)

}

type message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}
type DatabaseInterface interface {
	Get(key string) (string, error)
	Set(key string, value string) error

}

type DBRedis struct {
	client *redis.Client
	mu     sync.Mutex
}
type RateLimiter struct {
	DB DatabaseInterface
	mu sync.Mutex
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	m := message{
		Status: "OK",
		Body:   "success",
	}
	json.NewEncoder(w).Encode(m)
}

func perClientLimiter(next http.HandlerFunc) http.HandlerFunc {

	if err := godotenv.Load(); err != nil {
		slog.Error("Evnv.load", "message", "error loading .env file")
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	pong, err := rdb.Ping().Result()
	if err != nil {
		slog.Error("Redis.ping", "message", "error trying to connect to redis")
	}
	slog.Info("redis.NewClient", "message", "Connected to redis", "pong", pong)

	db := newRedisClient(rdb)

	max_ip, err := strconv.Atoi(os.Getenv("REQ_MAX_IP"))
	if err != nil {
		slog.Error("Redis.get", "message", "error trying to connect to redis")
	}
	db.Set("REQ_MAX_IP", ``+strconv.Itoa(max_ip))
	block_ip, err := strconv.Atoi(os.Getenv("BLOCK_TIME_IP"))
	if err != nil {
		slog.Error("Redis.get", "message", "error trying to connect to redis")
	}
	db.Set("BLOCK_TIME_IP", ``+strconv.Itoa(block_ip))
	type client struct {
		limiter    *rate.Limiter
		lastCalled time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			block_ip, err := db.Get("BLOCK_TIME_IP")
			if err != nil {
				slog.Error("Redis.get", "message", "error trying to connect to redis")
			}
			block_ip_int, _ := strconv.Atoi(block_ip)
			for ip, c := range clients {
				if time.Since(c.lastCalled) > time.Duration(block_ip_int)*time.Second {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		mu.Lock()
		max_ip, err := db.Get("REQ_MAX_IP")
		if err != nil {
			slog.Error("Redis.get", "message", "error trying to connect to redis")
		}
		max_ip_int, _ := strconv.Atoi(max_ip)
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(max_ip_int), 100)}
		}
		clients[ip].lastCalled = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			w.WriteHeader(http.StatusTooManyRequests)
			message := message{
				Status: "error",
				Body:   " you have reached the maximum number of requests or actions allowed within a certain time frame",
			}
			json.NewEncoder(w).Encode(message)
			return
		}
		mu.Unlock()
		next(w, r)
	}
}

func perTokenLimiter(next http.HandlerFunc) http.HandlerFunc {
	if err := godotenv.Load(); err != nil {
		slog.Error("Evnv.load", "message", "error loading .env file")
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	pong, err := rdb.Ping().Result()
	if err != nil {
		slog.Error("Redis.ping", "message", "error trying to connect to redis")
	}
	slog.Info("redis.NewClient", "message", "Connected to redis", "pong", pong)

	db := newRedisClient(rdb)

	max_auth, err := strconv.Atoi(os.Getenv("REQ_MAX_AUTH "))
	if err != nil {
		slog.Error("Redis.get", "message", "error trying to connect to redis")
	}
	db.Set("REQ_MAX_AUTH ", ``+strconv.Itoa(max_auth))
	block_auth, err := strconv.Atoi(os.Getenv("BLOCK_TIME_AUTH "))
	if err != nil {
		slog.Error("Redis.get", "message", "error trying to connect to redis")
	}
	db.Set("BLOCK_TIME_AUTH", ``+strconv.Itoa(block_auth))
	type token struct {
		limiter    *rate.Limiter
		lastCalled time.Time
	}

	var (
		mu     sync.Mutex
		tokens = make(map[string]*token)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			block_auth, err := db.Get("BLOCK_TIME_AUTH")
			if err != nil {
				slog.Error("Redis.get", "message", "error trying to connect to redis")
			}
			block_auth_int, _ := strconv.Atoi(block_auth)
			for auth, c := range tokens {
				if time.Since(c.lastCalled) > time.Duration(block_auth_int)*time.Second {
					delete(tokens, auth)
				}
			}
			mu.Unlock()
		}
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		auth := chi.URLParam(r, "API_KEY")

		mu.Lock()
		max_auth, err := db.Get("REQ_MAX_AUTH")
			if err != nil {
				slog.Error("Redis.get", "message", "error trying to connect to redis")
			}
		max_auth_int, _ := strconv.Atoi(max_auth)
		if _, found := tokens[auth]; !found {
			tokens[auth] = &token{limiter: rate.NewLimiter(rate.Limit(max_auth_int), 100)}
		}
		tokens[auth].lastCalled = time.Now()
		if !tokens[auth].limiter.Allow() {
			mu.Unlock()
			w.WriteHeader(http.StatusTooManyRequests)
			message := message{
				Status: "error",
				Body:   " you have reached the maximum number of requests or actions allowed within a certain time frame",
			}
			json.NewEncoder(w).Encode(message)
			return
		}
		mu.Unlock()
		next(w, r)
	}
}

func newRedisClient(client *redis.Client) DatabaseInterface {
	return &DBRedis{
		client: client,
	}
}

func (r *DBRedis) Get(key string) (string, error) {
	// gets a key
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.client.Get(key).Result()
}

func (r *DBRedis) Set(key string, value string) error {
	// sets a key
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.client.Set(key, value, time.Second*10).Err()
}