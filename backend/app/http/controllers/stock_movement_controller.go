package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// StockMovementController — read access to the stock ledger (PRD §8.14).
type StockMovementController struct{}

func NewStockMovementController() *StockMovementController {
	return &StockMovementController{}
}

func (r *StockMovementController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.StockMovement{}).
		With("Product").With("Warehouse").With("User")

	if pid := ctx.Request().QueryInt("product_id", 0); pid > 0 {
		query = query.Where("product_id", pid)
	}
	if wid := ctx.Request().QueryInt("warehouse_id", 0); wid > 0 {
		query = query.Where("warehouse_id", wid)
	}
	if t := ctx.Request().Query("type"); t != "" {
		query = query.Where("type", t)
	}
	if from := ctx.Request().Query("date_from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := ctx.Request().Query("date_to"); to != "" {
		query = query.Where("created_at <= ?", to+" 23:59:59")
	}

	var movements []models.StockMovement
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &movements, &total); err != nil {
		return serverError(ctx, "Failed to fetch stock movements")
	}

	return successMeta(ctx, "Stock movements retrieved successfully", movements, paginationMeta(page, perPage, total))
}
