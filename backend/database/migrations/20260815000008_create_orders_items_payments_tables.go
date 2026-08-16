package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260815000008CreateOrdersItemsPaymentsTables struct{}

func (r *M20260815000008CreateOrdersItemsPaymentsTables) Signature() string {
	return "20260815000008_create_orders_items_payments_tables"
}

// Up — PRD §8.19 (checkout), §8.18 (cart items), §8.20 (payment).
func (r *M20260815000008CreateOrdersItemsPaymentsTables) Up() error {
	if !facades.Schema().HasTable("orders") {
		if err := facades.Schema().Create("orders", func(table schema.Blueprint) {
			table.ID()
			table.String("order_number")
			table.UnsignedBigInteger("user_id").Comment("cashier")
			table.UnsignedBigInteger("customer_id").Nullable()
			table.UnsignedBigInteger("warehouse_id")
			table.Decimal("subtotal").Total(15).Places(2).Default(0)
			table.Decimal("discount").Total(15).Places(2).Default(0)
			table.Decimal("tax").Total(15).Places(2).Default(0)
			table.Decimal("grand_total").Total(15).Places(2).Default(0)
			table.Enum("status", []any{"pending", "completed", "cancelled"}).Default("completed")
			table.Text("notes").Nullable()
			table.Timestamp("ordered_at").Nullable()
			table.Timestamps()
			table.SoftDeletes()

			table.Unique("order_number")
			table.Index("user_id")
			table.Index("customer_id")
			table.Index("warehouse_id")
			table.Index("status")
			table.Foreign("user_id").References("id").On("users").RestrictOnDelete()
			table.Foreign("customer_id").References("id").On("customers").NullOnDelete()
			table.Foreign("warehouse_id").References("id").On("warehouses").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("order_items") {
		if err := facades.Schema().Create("order_items", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("order_id")
			table.UnsignedBigInteger("product_id")
			table.String("product_name").Comment("snapshot at sale time")
			table.String("sku").Nullable().Comment("snapshot at sale time")
			table.Decimal("unit_price").Total(15).Places(2)
			table.Integer("quantity")
			table.Decimal("discount").Total(15).Places(2).Default(0)
			table.Decimal("subtotal").Total(15).Places(2)
			table.Timestamps()

			table.Index("order_id")
			table.Index("product_id")
			table.Foreign("order_id").References("id").On("orders").CascadeOnDelete()
			table.Foreign("product_id").References("id").On("products").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("payments") {
		if err := facades.Schema().Create("payments", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("order_id")
			table.UnsignedBigInteger("payment_method_id")
			table.Decimal("amount").Total(15).Places(2).Comment("amount due allocated to this method")
			table.Decimal("amount_paid").Total(15).Places(2).Default(0)
			table.Decimal("change").Total(15).Places(2).Default(0)
			table.String("reference").Nullable()
			table.Enum("status", []any{"pending", "paid", "failed"}).Default("paid")
			table.Timestamp("paid_at").Nullable()
			table.Timestamps()

			table.Index("order_id")
			table.Index("payment_method_id")
			table.Foreign("order_id").References("id").On("orders").CascadeOnDelete()
			table.Foreign("payment_method_id").References("id").On("payment_methods").RestrictOnDelete()
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260815000008CreateOrdersItemsPaymentsTables) Down() error {
	if err := facades.Schema().DropIfExists("payments"); err != nil {
		return err
	}
	if err := facades.Schema().DropIfExists("order_items"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("orders")
}
