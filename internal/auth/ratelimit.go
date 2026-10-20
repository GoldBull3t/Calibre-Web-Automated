package auth

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ipBucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter gere limites de requisições por IP utilizando o algoritmo Token Bucket.
type RateLimiter struct {
	mu         sync.Mutex
	rate       float64 // tokens por segundo
	burst      float64 // capacidade máxima
	buckets    map[string]*ipBucket
	cleanupInt time.Duration
}

// NewRateLimiter inicializa um novo RateLimiter.
func NewRateLimiter(rate, burst float64) *RateLimiter {
	rl := &RateLimiter{
		rate:       rate,
		burst:      burst,
		buckets:    make(map[string]*ipBucket),
		cleanupInt: 10 * time.Minute,
	}

	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[ip]
	if !exists {
		rl.buckets[ip] = &ipBucket{
			tokens:     rl.burst - 1,
			lastRefill: now,
		}
		return true
	}

	// Recarrega tokens baseado no tempo decorrido
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInt)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-15 * time.Minute)
		for ip, b := range rl.buckets {
			if b.lastRefill.Before(cutoff) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// ExtractIP extrai o IP real do cliente considerando cabeçalhos de proxy reverso.
func ExtractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// Middleware envolve um http.Handler limitando requisições abusivas.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ExtractIP(r)
		if !rl.allow(ip) {
			w.Header().Set("Retry-After", "5")
			http.Error(w, `{"error":"demasiadas tentativas, aguarde alguns segundos"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

