package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

// Role — RBAC role (PRD §8.2): admin | cashier | warehouse.
type Role struct {
	orm.Model
	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`

	Users []User `gorm:"foreignKey:RoleID" json:"users,omitempty"`
}

// User — application user (PRD §8.3).
type User struct {
	orm.Model
	Name        string           `gorm:"column:name" json:"name"`
	Email       string           `gorm:"column:email" json:"email"`
	Phone       string           `gorm:"column:phone" json:"phone"`
	Password    string           `gorm:"column:password" json:"-"`
	RoleID      uint             `gorm:"column:role_id" json:"role_id"`
	Status      string           `gorm:"column:status" json:"status"`
	LastLoginAt *carbon.DateTime `gorm:"column:last_login_at" json:"last_login_at"`
	orm.SoftDeletes

	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}
