package middleware

import (
	nethttp "net/http"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// roleMiddleware restricts a route to users whose role is in the allowed set
// (PRD §8.2). Must run after Authenticate so the token is already parsed.
type roleMiddleware struct {
	roles []string
}

// Role returns middleware allowing only the given roles.
func Role(roles ...string) http.Middleware {
	return &roleMiddleware{roles: roles}
}

func (m *roleMiddleware) Signature() string {
	return "auth:role:" + strings.Join(m.roles, ",")
}

func (m *roleMiddleware) Handle(ctx http.Context) {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil || user.ID == 0 {
		abortJson(ctx, nethttp.StatusUnauthorized, "Unauthenticated")
		return
	}

	var role models.Role
	if err := facades.Orm().Query().Where("id", user.RoleID).First(&role); err != nil {
		abortJson(ctx, nethttp.StatusForbidden, "You do not have access to this resource")
		return
	}

	for _, allowed := range m.roles {
		if role.Name == allowed {
			ctx.Request().Next()
			return
		}
	}

	abortJson(ctx, nethttp.StatusForbidden, "You do not have access to this resource")
}
