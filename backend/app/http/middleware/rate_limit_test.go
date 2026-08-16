package middleware

import (
	"testing"
	"time"
)

func TestFixedWindowLimiterAllowsUpToMax(t *testing.T) {
	l := newFixedWindowLimiter(3, time.Minute)
	now := time.Now()

	for i := 1; i <= 3; i++ {
		if allowed, _ := l.allow("1.2.3.4", now); !allowed {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if allowed, retryAfter := l.allow("1.2.3.4", now); allowed {
		t.Fatalf("4th request should be blocked")
	} else if retryAfter <= 0 {
		t.Fatalf("blocked request should report a positive Retry-After, got %d", retryAfter)
	}
}

func TestFixedWindowLimiterResetsAfterWindow(t *testing.T) {
	l := newFixedWindowLimiter(1, time.Minute)
	now := time.Now()

	if allowed, _ := l.allow("ip", now); !allowed {
		t.Fatalf("first request should be allowed")
	}
	if allowed, _ := l.allow("ip", now); allowed {
		t.Fatalf("second request in window should be blocked")
	}
	// Move past the window; the counter should reset.
	if allowed, _ := l.allow("ip", now.Add(time.Minute+time.Second)); !allowed {
		t.Fatalf("request after window reset should be allowed")
	}
}

func TestFixedWindowLimiterKeysAreIndependent(t *testing.T) {
	l := newFixedWindowLimiter(1, time.Minute)
	now := time.Now()

	if allowed, _ := l.allow("a", now); !allowed {
		t.Fatalf("key a first request should be allowed")
	}
	if allowed, _ := l.allow("b", now); !allowed {
		t.Fatalf("key b should have its own independent budget")
	}
	if allowed, _ := l.allow("a", now); allowed {
		t.Fatalf("key a second request should be blocked")
	}
}

func TestFixedWindowLimiterSweepDropsExpired(t *testing.T) {
	l := newFixedWindowLimiter(5, time.Minute)
	now := time.Now()

	l.allow("stale", now)
	l.sweep(now.Add(2 * time.Minute))

	l.mu.Lock()
	_, present := l.hits["stale"]
	l.mu.Unlock()
	if present {
		t.Fatalf("expired key should have been swept")
	}
}
