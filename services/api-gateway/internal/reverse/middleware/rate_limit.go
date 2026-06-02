package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/newfeed/community-news/services/api-gateway/internal/reverse/response"
)

type bucket struct {
	count int
	reset time.Time
}

func RateLimit(requestsPerSecond int) Middleware {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 20
	}
	mu := sync.Mutex{}
	buckets := map[string]bucket{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := limiterIP(r)
			now := time.Now()
			mu.Lock()
			state := buckets[ip]
			if now.After(state.reset) {
				state = bucket{reset: now.Add(time.Second)}
			}
			state.count++
			buckets[ip] = state
			limited := state.count > requestsPerSecond
			mu.Unlock()
			if limited {
				response.Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func limiterIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
