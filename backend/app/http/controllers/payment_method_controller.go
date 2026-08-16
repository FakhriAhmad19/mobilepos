package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// PaymentMethodController — lists tender types for the POS (PRD §8.20).
type PaymentMethodController struct{}

func NewPaymentMethodController() *PaymentMethodController {
	return &PaymentMethodController{}
}

func (r *PaymentMethodController) Index(ctx http.Context) http.Response {
	query := facades.Orm().Query().Model(&models.PaymentMethod{})
	if ctx.Request().Query("active") == "1" {
		query = query.Where("is_active", true)
	}

	var methods []models.PaymentMethod
	if err := query.OrderBy("id").Get(&methods); err != nil {
		return serverError(ctx, "Failed to fetch payment methods")
	}
	return success(ctx, "Payment methods retrieved successfully", methods)
}
