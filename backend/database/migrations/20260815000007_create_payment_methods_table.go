package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000007CreatePaymentMethodsTable struct{}

func (r *M20260815000007CreatePaymentMethodsTable) Signature() string {
	return "20260815000007_create_payment_methods_table"
}

// Up — PRD §8.20 (payment). Seeded with cash/qris/debit/credit/transfer.
func (r *M20260815000007CreatePaymentMethodsTable) Up() error {
	if facades.Schema().HasTable("payment_methods") {
		return nil
	}
	return facades.Schema().Create("payment_methods", func(table schema.Blueprint) {
		table.ID()
		table.String("code").Comment("cash | qris | debit | credit | transfer")
		table.String("name")
		table.Boolean("is_active").Default(true)
		table.Timestamps()
		table.Unique("code")
	})
}

func (r *M20260815000007CreatePaymentMethodsTable) Down() error {
	return facades.Schema().DropIfExists("payment_methods")
}
