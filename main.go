package main

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
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

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// request for redis

}

func perClientLimiter(next http.HandlerFunc) http.HandlerFunc {

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
			for ip, c := range clients {
				if time.Since(c.lastCalled) > 5*time.Minute {
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
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Every(time.Second), 100)}
		}
		clients[ip].lastCalled = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			w.WriteHeader(http.StatusTooManyRequests)
			message := message{
				Status: "error",
				Body:   "Too many requests",
			}
			json.NewEncoder(w).Encode(message)
			return
		}
		mu.Unlock()
		next(w, r)
	}
}

func perTokenLimiter(next http.HandlerFunc) http.HandlerFunc {
	// auth := chi.URLParam(next, "API_KEY")
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
			for auth, c := range tokens {
				if time.Since(c.lastCalled) > 5*time.Minute {
					delete(tokens, auth)
				}
			}
			mu.Unlock()
		}
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		auth := chi.URLParam(r, "API_KEY")

		mu.Lock()
		if _, found := tokens[auth]; !found {
			tokens[auth] = &token{limiter: rate.NewLimiter(rate.Every(time.Second), 100)}
		}
		tokens[auth].lastCalled = time.Now()
		if !tokens[auth].limiter.Allow() {
			mu.Unlock()
			w.WriteHeader(http.StatusTooManyRequests)
			message := message{
				Status: "error",
				Body:   "Too many requests",
			}
			json.NewEncoder(w).Encode(message)
			return
		}
		mu.Unlock()
		next(w, r)
	}
}
