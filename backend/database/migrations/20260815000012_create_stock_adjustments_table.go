package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000012CreateStockAdjustmentsTable struct{}

func (r *M20260815000012CreateStockAdjustmentsTable) Signature() string {
	return "20260815000012_create_stock_adjustments_table"
}

// Up — PRD §8.12 (stock correction). One row per adjusted product.
func (r *M20260815000012CreateStockAdjustmentsTable) Up() error {
	if facades.Schema().HasTable("stock_adjustments") {
		return nil
	}
	return facades.Schema().Create("stock_adjustments", func(table schema.Blueprint) {
		table.ID()
		table.String("reference_number")
		table.UnsignedBigInteger("warehouse_id")
		table.UnsignedBigInteger("product_id")
		table.UnsignedBigInteger("user_id")
		table.Integer("quantity_before")
		table.Integer("quantity_after")
		table.Integer("difference").Comment("quantity_after - quantity_before")
		table.String("reason").Nullable()
		table.Text("notes").Nullable()
		table.Timestamps()

		table.Unique("reference_number")
		table.Index("warehouse_id")
		table.Index("product_id")
		table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
		table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		table.Foreign("user_id").References("id").On("users").RestrictOnDelete()
	})
}

func (r *M20260815000012CreateStockAdjustmentsTable) Down() error {
	return facades.Schema().DropIfExists("stock_adjustments")
}
