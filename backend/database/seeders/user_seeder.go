package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// UserSeeder seeds one user per role with known demo credentials (PRD §8.3).
type UserSeeder struct{}

func (s *UserSeeder) Signature() string {
	return "UserSeeder"
}

func (s *UserSeeder) Run() error {
	// Resolve role IDs seeded by RoleSeeder.
	roleIDs := map[string]uint{}
	var roles []models.Role
	if err := facades.Orm().Query().Find(&roles); err != nil {
		return err
	}
	for _, role := range roles {
		roleIDs[role.Name] = role.ID
	}

	hashed, err := facades.Hash().Make("password")
	if err != nil {
		return err
	}

	users := []models.User{
		{Name: "Admin Kasirku", Email: "admin@kasirku.test", Phone: "081200000001", Password: hashed, RoleID: roleIDs["admin"], Status: "active"},
		{Name: "Kasir Satu", Email: "cashier@kasirku.test", Phone: "081200000002", Password: hashed, RoleID: roleIDs["cashier"], Status: "active"},
		{Name: "Gudang Satu", Email: "warehouse@kasirku.test", Phone: "081200000003", Password: hashed, RoleID: roleIDs["warehouse"], Status: "active"},
	}

	for _, user := range users {
		count, err := facades.Orm().Query().Model(&models.User{}).Where("email", user.Email).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := facades.Orm().Query().Create(&user); err != nil {
			return err
		}
	}

	return nil
}
