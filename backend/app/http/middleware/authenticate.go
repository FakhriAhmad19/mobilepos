package middleware

import (
	nethttp "net/http"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// authenticateMiddleware verifies the JWT bearer token on the request and loads
// it into the auth context so downstream handlers can call
// facades.Auth(ctx).User(...).
type authenticateMiddleware struct{}

// Authenticate returns the JWT authentication middleware.
func Authenticate() http.Middleware {
	return &authenticateMiddleware{}
}

func (m *authenticateMiddleware) Signature() string {
	return "auth:jwt"
}

func (m *authenticateMiddleware) Handle(ctx http.Context) {
	token := ctx.Request().Header("Authorization", "")
	if token == "" {
		abortJson(ctx, nethttp.StatusUnauthorized, "Authentication token is required")
		return
	}

	if _, err := facades.Auth(ctx).Parse(token); err != nil {
		abortJson(ctx, nethttp.StatusUnauthorized, "Invalid or expired token")
		return
	}

	ctx.Request().Next()
}

// abortJson writes the PRD error envelope and stops the middleware chain.
func abortJson(ctx http.Context, status int, message string) {
	_ = ctx.Response().Json(status, http.Json{
		"success": false,
		"message": message,
	}).Abort()
}
