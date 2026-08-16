package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000004CreateWarehousesTable struct{}

func (r *M20260815000004CreateWarehousesTable) Signature() string {
	return "20260815000004_create_warehouses_table"
}

// Up — PRD §8.8 (multi-warehouse support).
func (r *M20260815000004CreateWarehousesTable) Up() error {
	if facades.Schema().HasTable("warehouses") {
		return nil
	}
	return facades.Schema().Create("warehouses", func(table schema.Blueprint) {
		table.ID()
		table.String("code")
		table.String("name")
		table.String("address").Nullable()
		table.String("phone").Nullable()
		table.Boolean("is_default").Default(false)
		table.Enum("status", []any{"active", "inactive"}).Default("active")
		table.Timestamps()
		table.SoftDeletes()
		table.Unique("code")
	})
}

func (r *M20260815000004CreateWarehousesTable) Down() error {
	return facades.Schema().DropIfExists("warehouses")
}
