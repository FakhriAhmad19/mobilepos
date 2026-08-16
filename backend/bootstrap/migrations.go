package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		// Kasirku schema (PRD §12) — order matters for foreign keys.
		&migrations.M20260815000001CreateRolesAndUsersTables{},
		&migrations.M20260815000002CreateCategoriesAndUnitsTables{},
		&migrations.M20260815000003CreateSuppliersAndCustomersTables{},
		&migrations.M20260815000004CreateWarehousesTable{},
		&migrations.M20260815000005CreateProductsTable{},
		&migrations.M20260815000006CreateStocksTable{},
		&migrations.M20260815000007CreatePaymentMethodsTable{},
		&migrations.M20260815000008CreateOrdersItemsPaymentsTables{},
		&migrations.M20260815000009CreateStockMovementsTable{},
		&migrations.M20260815000010CreateStockInsTables{},
		&migrations.M20260815000011CreateStockOutsTables{},
		&migrations.M20260815000012CreateStockAdjustmentsTable{},
		&migrations.M20260815000013CreateStockOpnamesTables{},
		&migrations.M20260815000014CreateNotificationsAndAuditLogsTables{},
	}
}
