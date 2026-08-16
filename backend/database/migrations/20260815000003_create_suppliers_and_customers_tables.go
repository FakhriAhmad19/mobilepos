package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000003CreateSuppliersAndCustomersTables struct{}

func (r *M20260815000003CreateSuppliersAndCustomersTables) Signature() string {
	return "20260815000003_create_suppliers_and_customers_tables"
}

// Up — PRD §8.6 (suppliers) and §8.7 (customers).
func (r *M20260815000003CreateSuppliersAndCustomersTables) Up() error {
	if !facades.Schema().HasTable("suppliers") {
		if err := facades.Schema().Create("suppliers", func(table schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("contact_person").Nullable()
			table.String("phone").Nullable()
			table.String("email").Nullable()
			table.String("address").Nullable()
			table.Enum("status", []any{"active", "inactive"}).Default("active")
			table.Timestamps()
			table.SoftDeletes()
			table.Index("name")
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("customers") {
		if err := facades.Schema().Create("customers", func(table schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("phone").Nullable()
			table.String("email").Nullable()
			table.String("address").Nullable()
			table.Enum("status", []any{"active", "inactive"}).Default("active")
			table.Timestamps()
			table.SoftDeletes()
			table.Index("name")
			table.Index("phone")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260815000003CreateSuppliersAndCustomersTables) Down() error {
	if err := facades.Schema().DropIfExists("customers"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("suppliers")
}
