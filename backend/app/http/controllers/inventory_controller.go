package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// InventoryController — read access to stock levels (PRD §8.9).
type InventoryController struct{}

func NewInventoryController() *InventoryController {
	return &InventoryController{}
}

func (r *InventoryController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.Stock{}).
		With("Product").With("Product.Category").With("Warehouse")

	if wid := ctx.Request().QueryInt("warehouse_id", 0); wid > 0 {
		query = query.Where("warehouse_id", wid)
	}
	if pid := ctx.Request().QueryInt("product_id", 0); pid > 0 {
		query = query.Where("product_id", pid)
	}
	if s := ctx.Request().Query("search"); s != "" {
		like := "%" + s + "%"
		query = query.Where(
			"product_id IN (SELECT id FROM app_products WHERE name LIKE ? OR sku LIKE ?)",
			like, like,
		)
	}
	// Low-stock: current quantity at or below the product's minimum (PRD §8.24).
	if ctx.Request().Query("low_stock") == "1" {
		query = query.Where("quantity <= (SELECT minimum_stock FROM app_products WHERE app_products.id = app_stocks.product_id)")
	}

	var stocks []models.Stock
	var total int64
	if err := query.OrderBy("warehouse_id").OrderByDesc("id").Paginate(page, perPage, &stocks, &total); err != nil {
		return serverError(ctx, "Failed to fetch inventory")
	}

	data := make([]http.Json, 0, len(stocks))
	for _, s := range stocks {
		low := s.Product != nil && s.Quantity <= s.Product.MinimumStock
		data = append(data, http.Json{
			"id":              s.ID,
			"product_id":      s.ProductID,
			"warehouse_id":    s.WarehouseID,
			"quantity":        s.Quantity,
			"reserved":        s.Reserved,
			"available_stock": s.AvailableStock(),
			"is_low_stock":    low,
			"product":         s.Product,
			"warehouse":       s.Warehouse,
		})
	}

	return successMeta(ctx, "Inventory retrieved successfully", data, paginationMeta(page, perPage, total))
}
