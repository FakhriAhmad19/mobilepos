package models

import "github.com/goravel/framework/support/carbon"

// Notification — in-app notification (PRD §8.25).
type Notification struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	UserID    *uint            `gorm:"column:user_id" json:"user_id"`
	Type      string           `gorm:"column:type" json:"type"`
	Title     string           `gorm:"column:title" json:"title"`
	Message   string           `gorm:"column:message" json:"message"`
	Data      map[string]any   `gorm:"column:data;serializer:json" json:"data"`
	ReadAt    *carbon.DateTime `gorm:"column:read_at" json:"read_at"`
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AuditLog — record of a significant action (PRD §8.26).
type AuditLog struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	UserID    *uint            `gorm:"column:user_id" json:"user_id"`
	Action    string           `gorm:"column:action" json:"action"`
	Module    string           `gorm:"column:module" json:"module"`
	Reference string           `gorm:"column:reference" json:"reference"`
	OldData   map[string]any   `gorm:"column:old_data;serializer:json" json:"old_data"`
	NewData   map[string]any   `gorm:"column:new_data;serializer:json" json:"new_data"`
	IPAddress string           `gorm:"column:ip_address" json:"ip_address"`
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
