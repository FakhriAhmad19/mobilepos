package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000001CreateRolesAndUsersTables struct{}

func (r *M20260815000001CreateRolesAndUsersTables) Signature() string {
	return "20260815000001_create_roles_and_users_tables"
}

// Up — PRD §8.2 (RBAC roles) and §8.3 (users).
func (r *M20260815000001CreateRolesAndUsersTables) Up() error {
	if !facades.Schema().HasTable("roles") {
		if err := facades.Schema().Create("roles", func(table schema.Blueprint) {
			table.ID()
			table.String("name").Comment("admin | cashier | warehouse")
			table.String("description").Nullable()
			table.Timestamps()
			table.Unique("name")
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("users") {
		if err := facades.Schema().Create("users", func(table schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email")
			table.String("phone").Nullable()
			table.String("password")
			table.UnsignedBigInteger("role_id")
			table.Enum("status", []any{"active", "inactive"}).Default("active")
			table.Timestamp("last_login_at").Nullable()
			table.Timestamps()
			table.SoftDeletes()

			table.Unique("email")
			table.Index("role_id")
			table.Foreign("role_id").References("id").On("roles").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260815000001CreateRolesAndUsersTables) Down() error {
	if err := facades.Schema().DropIfExists("users"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("roles")
}
