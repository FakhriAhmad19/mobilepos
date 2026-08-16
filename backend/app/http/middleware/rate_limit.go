package middleware

import (
	nethttp "net/http"
	"strconv"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/http"
)

// fixedWindowLimiter is a process-local, per-key fixed-window rate limiter.
// It intentionally avoids any external store (Redis, etc.) — a single in-memory
// counter per key is enough to blunt brute-force and abuse against sensitive
// endpoints such as /auth/login on one API instance.
type fixedWindowLimiter struct {
	mu     sync.Mutex
	hits   map[string]*windowCounter
	max    int
	window time.Duration
}

type windowCounter struct {
	count int
	reset time.Time
}

func newFixedWindowLimiter(max int, window time.Duration) *fixedWindowLimiter {
	return &fixedWindowLimiter{
		hits:   make(map[string]*windowCounter),
		max:    max,
		window: window,
	}
}

// allow records a hit for key at time now and reports whether it is within the
// limit, plus the number of whole seconds until the current window resets.
func (l *fixedWindowLimiter) allow(key string, now time.Time) (allowed bool, retryAfter int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Opportunistically drop expired keys so the map cannot grow unbounded
	// under a stream of distinct client IPs.
	if len(l.hits) > 1024 {
		l.sweep(now)
	}

	c, ok := l.hits[key]
	if !ok || now.After(c.reset) {
		l.hits[key] = &windowCounter{count: 1, reset: now.Add(l.window)}
		return true, int(l.window.Seconds())
	}

	c.count++
	retryAfter = int(c.reset.Sub(now).Seconds()) + 1
	if c.count > l.max {
		return false, retryAfter
	}
	return true, retryAfter
}

// sweep removes expired counters. Callers must hold l.mu.
func (l *fixedWindowLimiter) sweep(now time.Time) {
	for k, c := range l.hits {
		if now.After(c.reset) {
			delete(l.hits, k)
		}
	}
}

// throttleMiddleware limits each client IP to `max` requests per `window` for
// all routes registered with the same middleware instance.
type throttleMiddleware struct {
	name    string
	limiter *fixedWindowLimiter
}

// Throttle returns middleware allowing at most `max` requests per `window` per
// client IP. Requests over the limit are rejected with HTTP 429 and a
// Retry-After header. The returned middleware carries its own limiter, so keep
// a single instance per protected endpoint (register it once at route setup).
func Throttle(name string, max int, window time.Duration) http.Middleware {
	return &throttleMiddleware{
		name:    name,
		limiter: newFixedWindowLimiter(max, window),
	}
}

func (m *throttleMiddleware) Signature() string {
	return "throttle:" + m.name
}

func (m *throttleMiddleware) Handle(ctx http.Context) {
	key := m.name + "|" + ctx.Request().Ip()
	allowed, retryAfter := m.limiter.allow(key, time.Now())
	if !allowed {
		ctx.Response().Header("Retry-After", strconv.Itoa(retryAfter))
		abortJson(ctx, nethttp.StatusTooManyRequests, "Too many requests, please try again later")
		return
	}
	ctx.Request().Next()
}
