package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000011CreateStockOutsTables struct{}

func (r *M20260815000011CreateStockOutsTables) Signature() string {
	return "20260815000011_create_stock_outs_tables"
}

// Up — PRD §8.11 (non-sale stock removal: damaged, sample, internal, manual).
func (r *M20260815000011CreateStockOutsTables) Up() error {
	if !facades.Schema().HasTable("stock_outs") {
		if err := facades.Schema().Create("stock_outs", func(table schema.Blueprint) {
			table.ID()
			table.String("reference_number")
			table.UnsignedBigInteger("warehouse_id")
			table.UnsignedBigInteger("user_id")
			table.Date("date")
			table.Enum("reason", []any{"damaged", "sample", "internal", "manual"}).Default("manual")
			table.Enum("status", []any{"draft", "completed"}).Default("completed")
			table.Text("notes").Nullable()
			table.Timestamps()

			table.Unique("reference_number")
			table.Index("warehouse_id")
			table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("user_id").References("id").On("users").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("stock_out_items") {
		if err := facades.Schema().Create("stock_out_items", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("stock_out_id")
			table.UnsignedBigInteger("product_id")
			table.Integer("quantity")
			table.Timestamps()

			table.Index("stock_out_id")
			table.Index("product_id")
			table.Foreign("stock_out_id").References("id").On("stock_outs").CascadeOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260815000011CreateStockOutsTables) Down() error {
	if err := facades.Schema().DropIfExists("stock_out_items"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("stock_outs")
}
