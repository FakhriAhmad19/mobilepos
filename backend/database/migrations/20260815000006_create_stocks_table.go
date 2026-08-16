package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000006CreateStocksTable struct{}

func (r *M20260815000006CreateStocksTable) Signature() string {
	return "20260815000006_create_stocks_table"
}

// Up — PRD §8.9 (inventory). One row per product per warehouse.
// available_stock = quantity - reserved (computed at query time).
func (r *M20260815000006CreateStocksTable) Up() error {
	if facades.Schema().HasTable("stocks") {
		return nil
	}
	return facades.Schema().Create("stocks", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("product_id")
		table.UnsignedBigInteger("warehouse_id")
		table.Integer("quantity").Default(0).Comment("current stock")
		table.Integer("reserved").Default(0).Comment("reserved stock")
		table.Timestamps()

		table.Unique("product_id", "warehouse_id")
		table.Index("warehouse_id")
		table.Foreign("product_id").References("id").On("products").CascadeOnDelete()
		table.Foreign("warehouse_id").References("id").On("warehouses").CascadeOnDelete()
	})
}

func (r *M20260815000006CreateStocksTable) Down() error {
	return facades.Schema().DropIfExists("stocks")
}
