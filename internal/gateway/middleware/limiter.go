package middleware

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

func (m *Middleware) Limiter(next http.Handler) http.Handler {
	if !m.cfg.RateLimiter.Enabled {
		return next
	}

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realip.FromRequest(r)
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(m.cfg.RateLimiter.RPS), m.cfg.RateLimiter.Burst),
			}
		}
		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			helper.RateLimitExceededResponse(w, "Too many requests")
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
