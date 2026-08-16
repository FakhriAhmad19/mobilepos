package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000005CreateProductsTable struct{}

func (r *M20260815000005CreateProductsTable) Signature() string {
	return "20260815000005_create_products_table"
}

// Up — PRD §8.4 (products). Depends on categories & units.
func (r *M20260815000005CreateProductsTable) Up() error {
	if facades.Schema().HasTable("products") {
		return nil
	}
	return facades.Schema().Create("products", func(table schema.Blueprint) {
		table.ID()
		table.String("sku")
		table.String("barcode").Nullable()
		table.String("name")
		table.UnsignedBigInteger("category_id")
		table.UnsignedBigInteger("unit_id")
		table.Decimal("purchase_price").Total(15).Places(2).Default(0)
		table.Decimal("selling_price").Total(15).Places(2).Default(0)
		table.UnsignedInteger("minimum_stock").Default(0)
		table.String("image").Nullable()
		table.Text("description").Nullable()
		table.Enum("status", []any{"active", "inactive"}).Default("active")
		table.Timestamps()
		table.SoftDeletes()

		table.Unique("sku")
		table.Index("barcode")
		table.Index("name")
		table.Index("category_id")
		table.Index("unit_id")
		table.Foreign("category_id").References("id").On("categories").RestrictOnDelete()
		table.Foreign("unit_id").References("id").On("units").RestrictOnDelete()
	})
}

func (r *M20260815000005CreateProductsTable) Down() error {
	return facades.Schema().DropIfExists("products")
}
