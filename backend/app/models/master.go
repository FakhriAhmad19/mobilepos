package models

import "github.com/goravel/framework/database/orm"

// Category — product category (PRD §8.5).
type Category struct {
	orm.Model
	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`
	Status      string `gorm:"column:status" json:"status"`
	orm.SoftDeletes

	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// Unit — unit of measure (PRD §7 unit management).
type Unit struct {
	orm.Model
	Name        string `gorm:"column:name" json:"name"`
	Code        string `gorm:"column:code" json:"code"`
	Description string `gorm:"column:description" json:"description"`
	orm.SoftDeletes

	Products []Product `gorm:"foreignKey:UnitID" json:"products,omitempty"`
}

// Supplier — goods supplier (PRD §8.6).
type Supplier struct {
	orm.Model
	Name          string `gorm:"column:name" json:"name"`
	ContactPerson string `gorm:"column:contact_person" json:"contact_person"`
	Phone         string `gorm:"column:phone" json:"phone"`
	Email         string `gorm:"column:email" json:"email"`
	Address       string `gorm:"column:address" json:"address"`
	Status        string `gorm:"column:status" json:"status"`
	orm.SoftDeletes
}

// Customer — retail customer (PRD §8.7).
type Customer struct {
	orm.Model
	Name    string `gorm:"column:name" json:"name"`
	Phone   string `gorm:"column:phone" json:"phone"`
	Email   string `gorm:"column:email" json:"email"`
	Address string `gorm:"column:address" json:"address"`
	Status  string `gorm:"column:status" json:"status"`
	orm.SoftDeletes
}

// Warehouse — stock location (PRD §8.8).
type Warehouse struct {
	orm.Model
	Code      string `gorm:"column:code" json:"code"`
	Name      string `gorm:"column:name" json:"name"`
	Address   string `gorm:"column:address" json:"address"`
	Phone     string `gorm:"column:phone" json:"phone"`
	IsDefault bool   `gorm:"column:is_default" json:"is_default"`
	Status    string `gorm:"column:status" json:"status"`
	orm.SoftDeletes

	Stocks []Stock `gorm:"foreignKey:WarehouseID" json:"stocks,omitempty"`
}
