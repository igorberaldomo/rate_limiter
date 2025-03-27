package main

import (
	"net"
	"net/http"
	"time"
	"encoding/json"
	"sync"
	"golang.org/x/time/rate"

)

func main() {
	
	http.Handle("/", perClientLimiter(handler))
	
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

func perClientLimiter(next func(writer http.ResponseWriter, request *http.Request)) http.Handler {

	type client struct {
		limiter *rate.Limiter
		lastCalled    time.Time
	}

	var (
		mu   sync.Mutex
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		mu.Lock()
		if _ , found :=clients[ip]; !found {
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
	})
}