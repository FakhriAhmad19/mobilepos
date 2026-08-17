package middleware

import (
	nethttp "net/http"
	"strings"

	"github.com/goravel/framework/contracts/http"
)

// corsMiddleware adds permissive CORS headers to the API surface so browser
// clients (the Expo web build, and any web front-end) can call the API from a
// different origin/port. It is intentionally self-contained rather than relying
// on the gin driver's built-in CORS, whose config lookup is brittle across
// config value types. Credentials are not used (JWT travels in the
// Authorization header), so a wildcard origin is safe.
type corsMiddleware struct{}

// Cors returns the API CORS middleware.
func Cors() http.Middleware {
	return &corsMiddleware{}
}

func (m *corsMiddleware) Signature() string {
	return "cors"
}

func (m *corsMiddleware) Handle(ctx http.Context) {
	req := ctx.Request()

	// Only the API needs CORS; leave everything else untouched.
	if strings.HasPrefix(strings.TrimPrefix(req.Path(), "/"), "api/") {
		res := ctx.Response()
		res.Header("Access-Control-Allow-Origin", "*")
		res.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		res.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		res.Header("Access-Control-Max-Age", "86400")

		// Answer preflight requests directly — there is no OPTIONS route.
		if req.Method() == nethttp.MethodOptions {
			req.Abort(nethttp.StatusNoContent)
			return
		}
	}

	req.Next()
}
