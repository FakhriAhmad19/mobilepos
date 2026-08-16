package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

// Stock movement types (PRD §8.14).
const (
	MovementStockIn     = "STOCK_IN"
	MovementStockOut    = "STOCK_OUT"
	MovementSale        = "SALE"
	MovementReturn      = "RETURN"
	MovementAdjustment  = "ADJUSTMENT"
	MovementTransferIn  = "TRANSFER_IN"
	MovementTransferOut = "TRANSFER_OUT"
)

// StockMovement — immutable ledger row for every stock change (PRD §8.14).
type StockMovement struct {
	ID             uint             `gorm:"primaryKey" json:"id"`
	ProductID      uint             `gorm:"column:product_id" json:"product_id"`
	WarehouseID    uint             `gorm:"column:warehouse_id" json:"warehouse_id"`
	Type           string           `gorm:"column:type" json:"type"`
	Quantity       int              `gorm:"column:quantity" json:"quantity"`
	QuantityBefore int              `gorm:"column:quantity_before" json:"quantity_before"`
	QuantityAfter  int              `gorm:"column:quantity_after" json:"quantity_after"`
	ReferenceType  string           `gorm:"column:reference_type" json:"reference_type"`
	ReferenceID    *uint            `gorm:"column:reference_id" json:"reference_id"`
	UserID         *uint            `gorm:"column:user_id" json:"user_id"`
	Notes          string           `gorm:"column:notes" json:"notes"`
	CreatedAt      *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`

	Product   *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// StockIn — goods receiving header (PRD §8.10).
type StockIn struct {
	orm.Model
	ReferenceNumber string           `gorm:"column:reference_number" json:"reference_number"`
	SupplierID      *uint            `gorm:"column:supplier_id" json:"supplier_id"`
	WarehouseID     uint             `gorm:"column:warehouse_id" json:"warehouse_id"`
	UserID          uint             `gorm:"column:user_id" json:"user_id"`
	Date            *carbon.DateTime `gorm:"column:date" json:"date"`
	Status          string           `gorm:"column:status" json:"status"`
	Notes           string           `gorm:"column:notes" json:"notes"`

	Supplier  *Supplier     `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	Warehouse *Warehouse    `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	User      *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items     []StockInItem `gorm:"foreignKey:StockInID" json:"items,omitempty"`
}

// StockInItem — goods receiving line (PRD §8.10).
type StockInItem struct {
	orm.Model
	StockInID     uint    `gorm:"column:stock_in_id" json:"stock_in_id"`
	ProductID     uint    `gorm:"column:product_id" json:"product_id"`
	Quantity      int     `gorm:"column:quantity" json:"quantity"`
	PurchasePrice float64 `gorm:"column:purchase_price" json:"purchase_price"`
	Subtotal      float64 `gorm:"column:subtotal" json:"subtotal"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// StockOut — non-sale stock removal header (PRD §8.11).
type StockOut struct {
	orm.Model
	ReferenceNumber string           `gorm:"column:reference_number" json:"reference_number"`
	WarehouseID     uint             `gorm:"column:warehouse_id" json:"warehouse_id"`
	UserID          uint             `gorm:"column:user_id" json:"user_id"`
	Date            *carbon.DateTime `gorm:"column:date" json:"date"`
	Reason          string           `gorm:"column:reason" json:"reason"`
	Status          string           `gorm:"column:status" json:"status"`
	Notes           string           `gorm:"column:notes" json:"notes"`

	Warehouse *Warehouse     `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items     []StockOutItem `gorm:"foreignKey:StockOutID" json:"items,omitempty"`
}

// StockOutItem — non-sale stock removal line (PRD §8.11).
type StockOutItem struct {
	orm.Model
	StockOutID uint `gorm:"column:stock_out_id" json:"stock_out_id"`
	ProductID  uint `gorm:"column:product_id" json:"product_id"`
	Quantity   int  `gorm:"column:quantity" json:"quantity"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// StockAdjustment — stock correction (PRD §8.12).
type StockAdjustment struct {
	orm.Model
	ReferenceNumber string `gorm:"column:reference_number" json:"reference_number"`
	WarehouseID     uint   `gorm:"column:warehouse_id" json:"warehouse_id"`
	ProductID       uint   `gorm:"column:product_id" json:"product_id"`
	UserID          uint   `gorm:"column:user_id" json:"user_id"`
	QuantityBefore  int    `gorm:"column:quantity_before" json:"quantity_before"`
	QuantityAfter   int    `gorm:"column:quantity_after" json:"quantity_after"`
	Difference      int    `gorm:"column:difference" json:"difference"`
	Reason          string `gorm:"column:reason" json:"reason"`
	Notes           string `gorm:"column:notes" json:"notes"`

	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Product   *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// StockOpname — physical stock count header (PRD §8.13).
type StockOpname struct {
	orm.Model
	ReferenceNumber string           `gorm:"column:reference_number" json:"reference_number"`
	WarehouseID     uint             `gorm:"column:warehouse_id" json:"warehouse_id"`
	UserID          uint             `gorm:"column:user_id" json:"user_id"`
	Status          string           `gorm:"column:status" json:"status"`
	Notes           string           `gorm:"column:notes" json:"notes"`
	StartedAt       *carbon.DateTime `gorm:"column:started_at" json:"started_at"`
	CompletedAt     *carbon.DateTime `gorm:"column:completed_at" json:"completed_at"`

	Warehouse *Warehouse        `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	User      *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items     []StockOpnameItem `gorm:"foreignKey:StockOpnameID" json:"items,omitempty"`
}

// StockOpnameItem — counted line (PRD §8.13).
type StockOpnameItem struct {
	orm.Model
	StockOpnameID uint `gorm:"column:stock_opname_id" json:"stock_opname_id"`
	ProductID     uint `gorm:"column:product_id" json:"product_id"`
	SystemStock   int  `gorm:"column:system_stock" json:"system_stock"`
	PhysicalStock int  `gorm:"column:physical_stock" json:"physical_stock"`
	Difference    int  `gorm:"column:difference" json:"difference"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
