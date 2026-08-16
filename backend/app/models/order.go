package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

// PaymentMethod — accepted tender type (PRD §8.20).
type PaymentMethod struct {
	orm.Model
	Code     string `gorm:"column:code" json:"code"`
	Name     string `gorm:"column:name" json:"name"`
	IsActive bool   `gorm:"column:is_active" json:"is_active"`
}

// Order — a completed sale / POS transaction (PRD §8.19, §8.21).
type Order struct {
	orm.Model
	OrderNumber string           `gorm:"column:order_number" json:"order_number"`
	UserID      uint             `gorm:"column:user_id" json:"user_id"`
	CustomerID  *uint            `gorm:"column:customer_id" json:"customer_id"`
	WarehouseID uint             `gorm:"column:warehouse_id" json:"warehouse_id"`
	Subtotal    float64          `gorm:"column:subtotal" json:"subtotal"`
	Discount    float64          `gorm:"column:discount" json:"discount"`
	Tax         float64          `gorm:"column:tax" json:"tax"`
	GrandTotal  float64          `gorm:"column:grand_total" json:"grand_total"`
	Status      string           `gorm:"column:status" json:"status"`
	Notes       string           `gorm:"column:notes" json:"notes"`
	OrderedAt   *carbon.DateTime `gorm:"column:ordered_at" json:"ordered_at"`
	orm.SoftDeletes

	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Customer   *Customer   `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Warehouse  *Warehouse  `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
	Payments   []Payment   `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
}

// OrderItem — a line in an order (PRD §8.18). Product name/SKU are snapshotted.
type OrderItem struct {
	orm.Model
	OrderID     uint    `gorm:"column:order_id" json:"order_id"`
	ProductID   uint    `gorm:"column:product_id" json:"product_id"`
	ProductName string  `gorm:"column:product_name" json:"product_name"`
	SKU         string  `gorm:"column:sku" json:"sku"`
	UnitPrice   float64 `gorm:"column:unit_price" json:"unit_price"`
	Quantity    int     `gorm:"column:quantity" json:"quantity"`
	Discount    float64 `gorm:"column:discount" json:"discount"`
	Subtotal    float64 `gorm:"column:subtotal" json:"subtotal"`

	Order   *Order   `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// Payment — tender applied to an order (PRD §8.20).
type Payment struct {
	orm.Model
	OrderID         uint             `gorm:"column:order_id" json:"order_id"`
	PaymentMethodID uint             `gorm:"column:payment_method_id" json:"payment_method_id"`
	Amount          float64          `gorm:"column:amount" json:"amount"`
	AmountPaid      float64          `gorm:"column:amount_paid" json:"amount_paid"`
	Change          float64          `gorm:"column:change" json:"change"`
	Reference       string           `gorm:"column:reference" json:"reference"`
	Status          string           `gorm:"column:status" json:"status"`
	PaidAt          *carbon.DateTime `gorm:"column:paid_at" json:"paid_at"`

	Order         *Order         `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	PaymentMethod *PaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"payment_method,omitempty"`
}
