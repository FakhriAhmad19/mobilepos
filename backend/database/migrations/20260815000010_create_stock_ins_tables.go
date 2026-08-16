package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000010CreateStockInsTables struct{}

func (r *M20260815000010CreateStockInsTables) Signature() string {
	return "20260815000010_create_stock_ins_tables"
}

// Up — PRD §8.10 (goods receiving). Header + line items.
func (r *M20260815000010CreateStockInsTables) Up() error {
	if !facades.Schema().HasTable("stock_ins") {
		if err := facades.Schema().Create("stock_ins", func(table schema.Blueprint) {
			table.ID()
			table.String("reference_number")
			table.UnsignedBigInteger("supplier_id").Nullable()
			table.UnsignedBigInteger("warehouse_id")
			table.UnsignedBigInteger("user_id")
			table.Date("date")
			table.Enum("status", []any{"draft", "completed"}).Default("completed")
			table.Text("notes").Nullable()
			table.Timestamps()

			table.Unique("reference_number")
			table.Index("supplier_id")
			table.Index("warehouse_id")
			table.Foreign("supplier_id").References("id").On("suppliers").NullOnDelete()
			table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
			table.Foreign("user_id").References("id").On("users").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("stock_in_items") {
		if err := facades.Schema().Create("stock_in_items", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("stock_in_id")
			table.UnsignedBigInteger("product_id")
			table.Integer("quantity")
			table.Decimal("purchase_price").Total(15).Places(2).Default(0)
			table.Decimal("subtotal").Total(15).Places(2).Default(0)
			table.Timestamps()

			table.Index("stock_in_id")
			table.Index("product_id")
			table.Foreign("stock_in_id").References("id").On("stock_ins").CascadeOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260815000010CreateStockInsTables) Down() error {
	if err := facades.Schema().DropIfExists("stock_in_items"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("stock_ins")
}
