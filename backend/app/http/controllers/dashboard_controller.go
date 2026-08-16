package controllers

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// DashboardController — summary KPIs and charts (PRD §8.23).
type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

type dashboardSummary struct {
	TotalProducts     int64   `gorm:"column:total_products" json:"total_products"`
	TotalStock        int64   `gorm:"column:total_stock" json:"total_stock"`
	LowStock          int64   `gorm:"column:low_stock" json:"low_stock"`
	TodaySales        float64 `gorm:"column:today_sales" json:"today_sales"`
	TodayTransactions int64   `gorm:"column:today_transactions" json:"today_transactions"`
}

type salesPoint struct {
	Date  string  `gorm:"column:date" json:"date"`
	Total float64 `gorm:"column:total" json:"total"`
	Count int64   `gorm:"column:count" json:"count"`
}

type topProduct struct {
	ProductID   uint    `gorm:"column:product_id" json:"product_id"`
	ProductName string  `gorm:"column:product_name" json:"product_name"`
	Qty         int64   `gorm:"column:qty" json:"qty"`
	Revenue     float64 `gorm:"column:revenue" json:"revenue"`
}

func (r *DashboardController) Show(ctx http.Context) http.Response {
	var summary dashboardSummary
	if err := facades.Orm().Query().Raw(`
		SELECT
			(SELECT COUNT(*) FROM app_products WHERE deleted_at IS NULL) AS total_products,
			(SELECT COALESCE(SUM(quantity),0) FROM app_stocks) AS total_stock,
			(SELECT COUNT(*) FROM app_stocks s
				WHERE s.quantity <= (SELECT minimum_stock FROM app_products p WHERE p.id = s.product_id)) AS low_stock,
			(SELECT COALESCE(SUM(grand_total),0) FROM app_orders
				WHERE status='completed' AND deleted_at IS NULL AND DATE(created_at)=CURDATE()) AS today_sales,
			(SELECT COUNT(*) FROM app_orders
				WHERE status='completed' AND deleted_at IS NULL AND DATE(created_at)=CURDATE()) AS today_transactions
	`).Scan(&summary); err != nil {
		return serverError(ctx, "Failed to load dashboard")
	}

	// Sales for the last 7 days (gaps filled with zero).
	var points []salesPoint
	if err := facades.Orm().Query().Raw(`
		SELECT DATE(created_at) AS date, COALESCE(SUM(grand_total),0) AS total, COUNT(*) AS count
		FROM app_orders
		WHERE status='completed' AND deleted_at IS NULL
		  AND created_at >= (CURDATE() - INTERVAL 6 DAY)
		GROUP BY DATE(created_at)
		ORDER BY date
	`).Scan(&points); err != nil {
		return serverError(ctx, "Failed to load sales chart")
	}
	chart := fillSalesSeries(points)

	// Initialise to a non-nil slice so an empty result serialises as [] (not
	// null), which clients iterate over safely.
	top := []topProduct{}
	if err := facades.Orm().Query().Raw(`
		SELECT oi.product_id, oi.product_name, SUM(oi.quantity) AS qty, SUM(oi.subtotal) AS revenue
		FROM app_order_items oi
		JOIN app_orders o ON o.id = oi.order_id
		WHERE o.status='completed' AND o.deleted_at IS NULL
		GROUP BY oi.product_id, oi.product_name
		ORDER BY qty DESC
		LIMIT 5
	`).Scan(&top); err != nil {
		return serverError(ctx, "Failed to load top products")
	}

	return success(ctx, "Dashboard retrieved successfully", http.Json{
		"summary":      summary,
		"sales_7_days": chart,
		"top_products": top,
	})
}

// fillSalesSeries returns exactly 7 entries (oldest→newest), inserting zeros for
// days without sales so the chart has a continuous axis.
func fillSalesSeries(points []salesPoint) []salesPoint {
	byDate := map[string]salesPoint{}
	for _, p := range points {
		key := p.Date
		if len(key) > 10 {
			key = key[:10]
		}
		byDate[key] = salesPoint{Date: key, Total: p.Total, Count: p.Count}
	}

	series := make([]salesPoint, 0, 7)
	for i := 6; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if p, ok := byDate[day]; ok {
			series = append(series, p)
		} else {
			series = append(series, salesPoint{Date: day})
		}
	}
	return series
}
