package controllers

import (
	nethttp "net/http"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"

	"goravel/app/models"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// Login authenticates a user and issues a JWT (PRD §8.1).
// POST /api/v1/auth/login
func (r *AuthController) Login(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"email":    "required|email",
		"password": "required",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	email := ctx.Request().Input("email")
	password := ctx.Request().Input("password")

	var user models.User
	_ = facades.Orm().Query().With("Role").Where("email", email).First(&user)
	if user.ID == 0 {
		return unauthorized(ctx, "Invalid email or password")
	}
	if !facades.Hash().Check(password, user.Password) {
		return unauthorized(ctx, "Invalid email or password")
	}
	if user.Status != "active" {
		return ctx.Response().Json(nethttp.StatusForbidden, http.Json{
			"success": false,
			"message": "Your account is inactive",
		})
	}

	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return serverError(ctx, "Could not issue token")
	}

	// Record the login time (PRD §8.3 last_login_at).
	_, _ = facades.Orm().Query().Model(&models.User{}).
		Where("id", user.ID).
		Update("last_login_at", carbon.NewDateTime(carbon.Now()))

	return ctx.Response().Success().Json(http.Json{
		"success": true,
		"message": "Login successful",
		"data": http.Json{
			"token":      token,
			"token_type": "Bearer",
			"user":       userResource(user),
		},
		"meta": http.Json{},
	})
}

// Me returns the currently authenticated user (PRD §8.1 profile).
// GET /api/v1/auth/me
func (r *AuthController) Me(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil || user.ID == 0 {
		return unauthorized(ctx, "Unauthenticated")
	}
	_ = facades.Orm().Query().With("Role").Where("id", user.ID).First(&user)

	return ctx.Response().Success().Json(http.Json{
		"success": true,
		"message": "User retrieved successfully",
		"data":    userResource(user),
		"meta":    http.Json{},
	})
}

// Refresh issues a new token for the current session (PRD §8.1).
// POST /api/v1/auth/refresh
func (r *AuthController) Refresh(ctx http.Context) http.Response {
	token, err := facades.Auth(ctx).Refresh()
	if err != nil {
		return unauthorized(ctx, "Token could not be refreshed, please log in again")
	}

	return ctx.Response().Success().Json(http.Json{
		"success": true,
		"message": "Token refreshed",
		"data": http.Json{
			"token":      token,
			"token_type": "Bearer",
		},
		"meta": http.Json{},
	})
}

// Logout invalidates the current token (PRD §8.1).
// POST /api/v1/auth/logout
func (r *AuthController) Logout(ctx http.Context) http.Response {
	if err := facades.Auth(ctx).Logout(); err != nil {
		return serverError(ctx, "Could not log out")
	}

	return ctx.Response().Success().Json(http.Json{
		"success": true,
		"message": "Logged out successfully",
		"data":    http.Json{},
		"meta":    http.Json{},
	})
}

// ChangePassword updates the authenticated user's password (PRD §8.1).
// POST /api/v1/auth/change-password
func (r *AuthController) ChangePassword(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"current_password": "required",
		"new_password":     "required|min_len:6",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil || user.ID == 0 {
		return unauthorized(ctx, "Unauthenticated")
	}

	if !facades.Hash().Check(ctx.Request().Input("current_password"), user.Password) {
		return unauthorized(ctx, "Current password is incorrect")
	}

	hashed, err := facades.Hash().Make(ctx.Request().Input("new_password"))
	if err != nil {
		return serverError(ctx, "Could not update password")
	}
	if _, err := facades.Orm().Query().Model(&models.User{}).
		Where("id", user.ID).Update("password", hashed); err != nil {
		return serverError(ctx, "Could not update password")
	}

	return ctx.Response().Success().Json(http.Json{
		"success": true,
		"message": "Password changed successfully",
		"data":    http.Json{},
		"meta":    http.Json{},
	})
}

// --- helpers -------------------------------------------------------------

func userResource(user models.User) http.Json {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}
	return http.Json{
		"id":            user.ID,
		"name":          user.Name,
		"email":         user.Email,
		"phone":         user.Phone,
		"role":          roleName,
		"status":        user.Status,
		"last_login_at": user.LastLoginAt,
	}
}

func unauthorized(ctx http.Context, message string) http.Response {
	return ctx.Response().Json(nethttp.StatusUnauthorized, http.Json{
		"success": false,
		"message": message,
	})
}

func validationFailed(ctx http.Context, errs map[string]map[string]string) http.Response {
	return ctx.Response().Json(nethttp.StatusUnprocessableEntity, http.Json{
		"success": false,
		"message": "Validation failed",
		"errors":  errs,
	})
}

func serverError(ctx http.Context, message string) http.Response {
	return ctx.Response().Json(nethttp.StatusInternalServerError, http.Json{
		"success": false,
		"message": message,
	})
}
