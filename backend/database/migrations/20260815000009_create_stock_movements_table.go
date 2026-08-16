package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000009CreateStockMovementsTable struct{}

func (r *M20260815000009CreateStockMovementsTable) Signature() string {
	return "20260815000009_create_stock_movements_table"
}

// Up — PRD §8.14. Every stock change writes an immutable movement row.
// reference_type/reference_id polymorphically link the source document
// (stock_in, stock_out, order, adjustment, opname, transfer).
func (r *M20260815000009CreateStockMovementsTable) Up() error {
	if facades.Schema().HasTable("stock_movements") {
		return nil
	}
	return facades.Schema().Create("stock_movements", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("product_id")
		table.UnsignedBigInteger("warehouse_id")
		table.Enum("type", []any{
			"STOCK_IN", "STOCK_OUT", "SALE", "RETURN",
			"ADJUSTMENT", "TRANSFER_IN", "TRANSFER_OUT",
		})
		table.Integer("quantity").Comment("signed delta applied to stock")
		table.Integer("quantity_before")
		table.Integer("quantity_after")
		table.String("reference_type").Nullable()
		table.UnsignedBigInteger("reference_id").Nullable()
		table.UnsignedBigInteger("user_id").Nullable()
		table.Text("notes").Nullable()
		table.Timestamp("created_at").UseCurrent()

		table.Index("product_id")
		table.Index("warehouse_id")
		table.Index("type")
		table.Index("reference_type", "reference_id")
		table.Foreign("product_id").References("id").On("products").CascadeOnDelete()
		table.Foreign("warehouse_id").References("id").On("warehouses").CascadeOnDelete()
		table.Foreign("user_id").References("id").On("users").NullOnDelete()
	})
}

func (r *M20260815000009CreateStockMovementsTable) Down() error {
	return facades.Schema().DropIfExists("stock_movements")
}
