package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000014CreateNotificationsAndAuditLogsTables struct{}

func (r *M20260815000014CreateNotificationsAndAuditLogsTables) Signature() string {
	return "20260815000014_create_notifications_and_audit_logs_tables"
}

// Up — PRD §8.25 (notifications) and §8.26 (audit log).
func (r *M20260815000014CreateNotificationsAndAuditLogsTables) Up() error {
	if !facades.Schema().HasTable("notifications") {
		if err := facades.Schema().Create("notifications", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id").Nullable().Comment("null = broadcast")
			table.String("type").Comment("low_stock | out_of_stock | sale | opname_done")
			table.String("title")
			table.Text("message").Nullable()
			table.Json("data").Nullable()
			table.Timestamp("read_at").Nullable()
			table.Timestamp("created_at").UseCurrent()

			table.Index("user_id")
			table.Index("type")
			table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("audit_logs") {
		if err := facades.Schema().Create("audit_logs", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id").Nullable()
			table.String("action").Comment("login | create | update | delete | ...")
			table.String("module").Comment("product | stock | transaction | ...")
			table.String("reference").Nullable()
			table.Json("old_data").Nullable()
			table.Json("new_data").Nullable()
			table.String("ip_address").Nullable()
			table.Timestamp("created_at").UseCurrent()

			table.Index("user_id")
			table.Index("module")
			table.Index("action")
			table.Foreign("user_id").References("id").On("users").NullOnDelete()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260815000014CreateNotificationsAndAuditLogsTables) Down() error {
	if err := facades.Schema().DropIfExists("audit_logs"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("notifications")
}
