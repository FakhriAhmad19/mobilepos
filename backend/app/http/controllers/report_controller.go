package controllers

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// ReportController — sales, product, inventory and cashier reports (PRD §8.24).
type ReportController struct{}

func NewReportController() *ReportController {
	return &ReportController{}
}

// dateRange reads ?date_from / ?date_to (YYYY-MM-DD), defaulting to the last
// 30 days. `to` is returned inclusive of the whole day.
func dateRange(ctx http.Context) (from, to string) {
	from = ctx.Request().Query("date_from")
	to = ctx.Request().Query("date_to")
	if from == "" {
		from = time.Now().AddDate(0, 0, -29).Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02")
	}
	return from, to + " 23:59:59"
}

// Sales — daily sales totals over a date range (PRD §8.24 Sales Report).
func (r *ReportController) Sales(ctx http.Context) http.Response {
	from, to := dateRange(ctx)

	var rows []salesPoint
	if err := facades.Orm().Query().Raw(`
		SELECT DATE(created_at) AS date, COALESCE(SUM(grand_total),0) AS total, COUNT(*) AS count
		FROM app_orders
		WHERE status='completed' AND deleted_at IS NULL AND created_at BETWEEN ? AND ?
		GROUP BY DATE(created_at)
		ORDER BY date
	`, from, to).Scan(&rows); err != nil {
		return serverError(ctx, "Failed to build sales report")
	}

	var totalSales float64
	var totalTxns int64
	for _, row := range rows {
		totalSales += row.Total
		totalTxns += row.Count
	}

	return successMeta(ctx, "Sales report generated", rows, http.Json{
		"date_from":    from,
		"date_to":      to,
		"total_sales":  totalSales,
		"total_orders": totalTxns,
	})
}

// Products — best or worst selling products (PRD §8.24 Product Report).
// ?sort=best|worst (default best).
func (r *ReportController) Products(ctx http.Context) http.Response {
	from, to := dateRange(ctx)
	order := "DESC"
	if ctx.Request().Query("sort") == "worst" {
		order = "ASC"
	}

	var rows []topProduct
	if err := facades.Orm().Query().Raw(`
		SELECT oi.product_id, oi.product_name, SUM(oi.quantity) AS qty, SUM(oi.subtotal) AS revenue
		FROM app_order_items oi
		JOIN app_orders o ON o.id = oi.order_id
		WHERE o.status='completed' AND o.deleted_at IS NULL AND o.created_at BETWEEN ? AND ?
		GROUP BY oi.product_id, oi.product_name
		ORDER BY qty `+order+`
		LIMIT 20
	`, from, to).Scan(&rows); err != nil {
		return serverError(ctx, "Failed to build product report")
	}
	return successMeta(ctx, "Product report generated", rows, http.Json{
		"date_from": from, "date_to": to, "sort": ctx.Request().Query("sort", "best"),
	})
}

// Inventory — current stock valuation and low-stock alerts (PRD §8.24).
func (r *ReportController) Inventory(ctx http.Context) http.Response {
	type invSummary struct {
		TotalStock  int64 `gorm:"column:total_stock" json:"total_stock"`
		LowStock    int64 `gorm:"column:low_stock" json:"low_stock"`
		OutOfStock  int64 `gorm:"column:out_of_stock" json:"out_of_stock"`
		SkuCount    int64 `gorm:"column:sku_count" json:"sku_count"`
	}
	var summary invSummary
	if err := facades.Orm().Query().Raw(`
		SELECT
			COALESCE(SUM(quantity),0) AS total_stock,
			COUNT(*) AS sku_count,
			SUM(CASE WHEN quantity <= (SELECT minimum_stock FROM app_products p WHERE p.id = s.product_id) THEN 1 ELSE 0 END) AS low_stock,
			SUM(CASE WHEN quantity <= 0 THEN 1 ELSE 0 END) AS out_of_stock
		FROM app_stocks s
	`).Scan(&summary); err != nil {
		return serverError(ctx, "Failed to build inventory report")
	}

	// Low-stock detail list.
	var low []models.Stock
	if err := facades.Orm().Query().Model(&models.Stock{}).With("Product").With("Warehouse").
		Where("quantity <= (SELECT minimum_stock FROM app_products WHERE app_products.id = app_stocks.product_id)").
		OrderBy("quantity").Limit(50).Get(&low); err != nil {
		return serverError(ctx, "Failed to load low-stock items")
	}

	items := make([]http.Json, 0, len(low))
	for _, s := range low {
		items = append(items, http.Json{
			"product_id": s.ProductID, "warehouse_id": s.WarehouseID,
			"quantity": s.Quantity, "available_stock": s.AvailableStock(),
			"product": s.Product, "warehouse": s.Warehouse,
		})
	}

	return success(ctx, "Inventory report generated", http.Json{
		"summary":         summary,
		"low_stock_items": items,
	})
}

// Cashiers — transactions and sales per cashier (PRD §8.24 Cashier Report).
func (r *ReportController) Cashiers(ctx http.Context) http.Response {
	from, to := dateRange(ctx)

	type cashierRow struct {
		UserID       uint    `gorm:"column:user_id" json:"user_id"`
		Name         string  `gorm:"column:name" json:"name"`
		Transactions int64   `gorm:"column:transactions" json:"transactions"`
		Sales        float64 `gorm:"column:sales" json:"sales"`
	}
	var rows []cashierRow
	if err := facades.Orm().Query().Raw(`
		SELECT o.user_id, u.name, COUNT(*) AS transactions, COALESCE(SUM(o.grand_total),0) AS sales
		FROM app_orders o
		JOIN app_users u ON u.id = o.user_id
		WHERE o.status='completed' AND o.deleted_at IS NULL AND o.created_at BETWEEN ? AND ?
		GROUP BY o.user_id, u.name
		ORDER BY sales DESC
	`, from, to).Scan(&rows); err != nil {
		return serverError(ctx, "Failed to build cashier report")
	}
	return successMeta(ctx, "Cashier report generated", rows, http.Json{
		"date_from": from, "date_to": to,
	})
}
