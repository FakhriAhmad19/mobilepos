package middleware

import (
	"github.com/goravel/framework/contracts/http"
)

// secureHeadersMiddleware attaches a baseline set of hardening response headers
// to every response (Phase 8 — security). The values are safe defaults for a
// JSON-only API that serves no HTML and embeds no third-party resources.
type secureHeadersMiddleware struct{}

// SecureHeaders returns middleware that sets baseline security headers.
func SecureHeaders() http.Middleware {
	return &secureHeadersMiddleware{}
}

func (m *secureHeadersMiddleware) Signature() string {
	return "secure-headers"
}

func (m *secureHeadersMiddleware) Handle(ctx http.Context) {
	res := ctx.Response()
	// Stop browsers from MIME-sniffing responses away from the declared type.
	res.Header("X-Content-Type-Options", "nosniff")
	// This API never needs to be framed; deny it outright (clickjacking).
	res.Header("X-Frame-Options", "DENY")
	// Do not leak the request URL (which can carry ids) via the Referer header.
	res.Header("Referrer-Policy", "no-referrer")
	// Disable the legacy, buggy XSS auditor rather than enable it (modern advice).
	res.Header("X-XSS-Protection", "0")
	// A JSON API loads nothing; lock the CSP all the way down.
	res.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

	ctx.Request().Next()
}
