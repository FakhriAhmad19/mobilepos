package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"

	"goravel/app/facades"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// Show reports service liveness and database connectivity.
// GET /api/v1/health
func (r *HealthController) Show(ctx http.Context) http.Response {
	dbStatus := "up"
	if _, err := facades.Orm().Query().Exec("SELECT 1"); err != nil {
		dbStatus = "down"
	}

	return ctx.Response().Success().Json(http.Json{
		"success": true,
		"message": "Service healthy",
		"data": http.Json{
			"service":   facades.Config().GetString("app.name", "KasirkuMobile"),
			"env":       facades.Config().GetString("app.env"),
			"database":  dbStatus,
			"timestamp": carbon.Now().ToDateTimeString(),
		},
		"meta": http.Json{},
	})
}
