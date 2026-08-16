package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260816000001AddReportIndexes struct{}

func (r *M20260816000001AddReportIndexes) Signature() string {
	return "20260816000001_add_report_indexes"
}

// Up adds a composite index on orders(status, created_at) (Phase 8 — perf).
//
// Every dashboard/report query and the order-history date filter scan
// `app_orders` with `status='completed' AND created_at BETWEEN ? AND ?`
// (sales, top-products, cashiers reports; orders list). Before this index the
// table had `status` alone, so the created_at range still meant a scan of every
// completed order. The composite lets MySQL seek straight to the status +
// date-range slice.
func (r *M20260816000001AddReportIndexes) Up() error {
	if facades.Schema().HasIndex("orders", indexOrdersStatusCreatedAt) {
		return nil
	}
	return facades.Schema().Table("orders", func(table schema.Blueprint) {
		table.Index("status", "created_at").Name(indexOrdersStatusCreatedAt)
	})
}

func (r *M20260816000001AddReportIndexes) Down() error {
	if !facades.Schema().HasIndex("orders", indexOrdersStatusCreatedAt) {
		return nil
	}
	return facades.Schema().Table("orders", func(table schema.Blueprint) {
		table.DropIndexByName(indexOrdersStatusCreatedAt)
	})
}

const indexOrdersStatusCreatedAt = "idx_orders_status_created_at"
