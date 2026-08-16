package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// RoleSeeder seeds the three RBAC roles (PRD §8.2).
type RoleSeeder struct{}

func (s *RoleSeeder) Signature() string {
	return "RoleSeeder"
}

func (s *RoleSeeder) Run() error {
	roles := []models.Role{
		{Name: "admin", Description: "Full system access"},
		{Name: "cashier", Description: "Point of sale operations"},
		{Name: "warehouse", Description: "Inventory management"},
	}

	for _, role := range roles {
		count, err := facades.Orm().Query().Model(&models.Role{}).Where("name", role.Name).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := facades.Orm().Query().Create(&role); err != nil {
			return err
		}
	}

	return nil
}
