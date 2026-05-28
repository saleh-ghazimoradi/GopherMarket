package middleware

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

type clientItem struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ConcurrentLimiter struct {
	mu      sync.RWMutex
	clients map[string]*clientItem
}

func (m *Middleware) Limiter(next http.Handler) http.Handler {
	if !m.cfg.RateLimiter.Enabled {
		return next
	}

	state := &ConcurrentLimiter{
		clients: make(map[string]*clientItem),
	}

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			state.mu.Lock()
			for ip, item := range state.clients {
				if time.Since(item.lastSeen) > 3*time.Minute {
					delete(state.clients, ip)
				}
			}
			state.mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realip.FromRequest(r)
		now := time.Now()

		state.mu.RLock()
		cl, exists := state.clients[ip]
		state.mu.RUnlock()

		if !exists {
			state.mu.Lock()
			cl, exists = state.clients[ip]
			if !exists {
				cl = &clientItem{
					limiter: rate.NewLimiter(rate.Limit(m.cfg.RateLimiter.RPS), m.cfg.RateLimiter.Burst),
				}
				state.clients[ip] = cl
			}
			state.mu.Unlock()
		}

		state.mu.Lock()
		cl.lastSeen = now
		allowed := cl.limiter.Allow()
		state.mu.Unlock()

		if !allowed {
			helper.RateLimitExceededResponse(w, "Too Many Requests")
			return
		}

		next.ServeHTTP(w, r)
	})
}
