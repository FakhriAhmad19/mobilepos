package models

import "github.com/goravel/framework/database/orm"

// Product — sellable item (PRD §8.4).
type Product struct {
	orm.Model
	SKU           string  `gorm:"column:sku" json:"sku"`
	Barcode       string  `gorm:"column:barcode" json:"barcode"`
	Name          string  `gorm:"column:name" json:"name"`
	CategoryID    uint    `gorm:"column:category_id" json:"category_id"`
	UnitID        uint    `gorm:"column:unit_id" json:"unit_id"`
	PurchasePrice float64 `gorm:"column:purchase_price" json:"purchase_price"`
	SellingPrice  float64 `gorm:"column:selling_price" json:"selling_price"`
	MinimumStock  int     `gorm:"column:minimum_stock" json:"minimum_stock"`
	Image         string  `gorm:"column:image" json:"image"`
	Description   string  `gorm:"column:description" json:"description"`
	Status        string  `gorm:"column:status" json:"status"`
	orm.SoftDeletes

	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Unit     *Unit     `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Stocks   []Stock   `gorm:"foreignKey:ProductID" json:"stocks,omitempty"`
}

// Stock — per-product, per-warehouse inventory level (PRD §8.9).
// AvailableStock (quantity - reserved) is computed, not stored.
type Stock struct {
	orm.Model
	ProductID   uint `gorm:"column:product_id" json:"product_id"`
	WarehouseID uint `gorm:"column:warehouse_id" json:"warehouse_id"`
	Quantity    int  `gorm:"column:quantity" json:"quantity"`
	Reserved    int  `gorm:"column:reserved" json:"reserved"`

	Product   *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

// AvailableStock returns quantity minus reserved (PRD §8.9 formula).
func (s Stock) AvailableStock() int {
	return s.Quantity - s.Reserved
}
