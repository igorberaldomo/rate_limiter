package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var RedisClient DatabaseInterface

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := godotenv.Load(".env"); err != nil {
		slog.Error("Env.load", "message", "error loading .env file")
	}
	RedisClient = newRedisClient(redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})).(*DBRedis)

	salvarEnvVars()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", perClientLimiter(handler))
	r.Get("/{API_KEY}", perTokenLimiter(handler))

	http.ListenAndServe(":8080", r)

}

func salvarEnvVars() {
	vars := []string{"REQ_MAX_IP", "BLOCK_TIME_IP", "REQ_MAX_AUTH", "BLOCK_TIME_AUTH"}
	for _, v := range vars {
		if os.Getenv(v) == "" {
			slog.Error("salvarEnvVars", "message", "error trying to recover "+v)
			return
		} else {
			t, err := strconv.Atoi(os.Getenv(v))
			if err != nil {
				slog.Error("salvarEnvVars", "message", "error trying to converting "+v, "error", err)
			}
			RedisClient.Set(v, t, 0)
		}
	}
}

type message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

type DatabaseInterface interface {
	Get(key string) (int, error)
	Set(key string, value int, time time.Duration) error
	Update(key string, value int) error
}

type DBRedis struct {
	Client *redis.Client
	Mu     sync.Mutex
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
	return func(w http.ResponseWriter, r *http.Request) {

		db := RedisClient.(*DBRedis)
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		max_ip, err := db.Get("REQ_MAX_IP")
		if err != nil {
			slog.Error("perClientLimiter2", "message", "error retrieng REQ_MAX_IP", "error", err)
		}
		block_time, err := db.Get("BLOCK_TIME_IP")
		if err != nil {
			slog.Error("perClientLimiter2", "message", "error retrieng BLOCK_TIME_IP", "error", err)
		}
		slog.Info("perClientLimiter", "ip", ip, "max_ip", max_ip, "block_time", block_time)
		found, err := db.Get(ip)
		if err != nil {
			db.Set(ip, 1, time.Second*time.Duration(block_time))
			next(w, r)
		} else {
			if found >= max_ip {
				w.WriteHeader(http.StatusTooManyRequests)
				message := message{
					Status: "error",
					Body:   " you have reached the maximum number of requests or actions allowed within a certain time frame",
				}
				json.NewEncoder(w).Encode(message)
				return
			} else {
				err = db.Update(ip, found+1)
				if err != nil {
					slog.Error("perClientLimiter2", "message", "updating access count", "error", err)
				}
				next(w, r)
			}
		}
	}
}

func perTokenLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := chi.URLParam(r, "API_KEY")
		db := RedisClient.(*DBRedis)
		max_auth, err := db.Get("REQ_MAX_AUTH")
		if err != nil {
			slog.Error("perTokenLimiter", "message", "error db.get REQ_MAX_AUTH", "error", err, "nax_auth", max_auth)
		}
		block_time, err := db.Get("BLOCK_TIME_AUTH")
		if err != nil {
			slog.Error("perTokenLimiter", "message", "error retrieng BLOCK_TIME_AUTH", "error", err)
		}
		found, err := db.Get(auth)
		if err != nil {
			err = db.Set(auth, 1, time.Second*time.Duration(block_time))
			if err != nil {
				slog.Error("perTokenLimiter", "message", "error db.set", "error", err)
			}
			next(w, r)
		} else {
			if found >= max_auth {
				w.WriteHeader(http.StatusTooManyRequests)
				message := message{
					Status: "error",
					Body:   " you have reached the maximum number of requests or actions allowed within a certain time frame",
				}
				json.NewEncoder(w).Encode(message)
				return
			} else {
				err = db.Update(auth, found+1)
				if err != nil {
					slog.Error("perTokenLimiter", "message", "updating access count", "error", err)
				}
				next(w, r)
			}
		}
	}
}

func newRedisClient(client *redis.Client) DatabaseInterface {
	return &DBRedis{
		Client: client,
		Mu:     sync.Mutex{},
	}
}

func (r *DBRedis) Get(key string) (int, error) {
	// gets a key
	r.Mu.Lock()
	defer r.Mu.Unlock()
	val, err := r.Client.Get(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func (r *DBRedis) Set(key string, value int, time time.Duration) error {
	// sets a key
	r.Mu.Lock()
	defer r.Mu.Unlock()
	return r.Client.Set(context.Background(), key, value, time).Err()
}

func (r *DBRedis) Update(key string, value int) error {
	// updates a key
	r.Mu.Lock()
	defer r.Mu.Unlock()
	result, err := r.Client.Set(context.Background(), key, value, redis.KeepTTL).Result()
	if err != nil {
		slog.Info("Update", "key", key, "value", value, "result", result, "error", err)
	}
	slog.Info("Update", "key", key, "value", value, "result", result)
	return nil
}
