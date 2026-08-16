package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000002CreateCategoriesAndUnitsTables struct{}

func (r *M20260815000002CreateCategoriesAndUnitsTables) Signature() string {
	return "20260815000002_create_categories_and_units_tables"
}

// Up — PRD §8.5 (categories) and unit management (§7).
func (r *M20260815000002CreateCategoriesAndUnitsTables) Up() error {
	if !facades.Schema().HasTable("categories") {
		if err := facades.Schema().Create("categories", func(table schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("description").Nullable()
			table.Enum("status", []any{"active", "inactive"}).Default("active")
			table.Timestamps()
			table.SoftDeletes()
			table.Index("name")
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("units") {
		if err := facades.Schema().Create("units", func(table schema.Blueprint) {
			table.ID()
			table.String("name").Comment("e.g. Pieces, Box, Kilogram")
			table.String("code").Comment("e.g. pcs, box, kg")
			table.String("description").Nullable()
			table.Timestamps()
			table.SoftDeletes()
			table.Unique("code")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260815000002CreateCategoriesAndUnitsTables) Down() error {
	if err := facades.Schema().DropIfExists("units"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("categories")
}
