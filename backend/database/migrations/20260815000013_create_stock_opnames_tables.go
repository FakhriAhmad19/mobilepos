package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000013CreateStockOpnamesTables struct{}

func (r *M20260815000013CreateStockOpnamesTables) Signature() string {
	return "20260815000013_create_stock_opnames_tables"
}

// Up — PRD §8.13 (physical stock count / opname). Header + counted items.
func (r *M20260815000013CreateStockOpnamesTables) Up() error {
	if !facades.Schema().HasTable("stock_opnames") {
		if err := facades.Schema().Create("stock_opnames", func(table schema.Blueprint) {
			table.ID()
			table.String("reference_number")
			table.UnsignedBigInteger("warehouse_id")
			table.UnsignedBigInteger("user_id")
			table.Enum("status", []any{"draft", "in_progress", "completed"}).Default("draft")
			table.Text("notes").Nullable()
			table.Timestamp("started_at").Nullable()
			table.Timestamp("completed_at").Nullable()
			table.Timestamps()

			table.Unique("reference_number")
			table.Index("warehouse_id")
			table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("user_id").References("id").On("users").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("stock_opname_items") {
		if err := facades.Schema().Create("stock_opname_items", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("stock_opname_id")
			table.UnsignedBigInteger("product_id")
			table.Integer("system_stock")
			table.Integer("physical_stock")
			table.Integer("difference").Comment("physical_stock - system_stock")
			table.Timestamps()

			table.Index("stock_opname_id")
			table.Index("product_id")
			table.Foreign("stock_opname_id").References("id").On("stock_opnames").CascadeOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260815000013CreateStockOpnamesTables) Down() error {
	if err := facades.Schema().DropIfExists("stock_opname_items"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("stock_opnames")
}
