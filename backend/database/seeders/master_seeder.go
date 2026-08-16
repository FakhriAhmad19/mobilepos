package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// MasterSeeder seeds categories, units, warehouses and payment methods
// (PRD §8.5, §8.8, §8.20).
type MasterSeeder struct{}

func (s *MasterSeeder) Signature() string {
	return "MasterSeeder"
}

func (s *MasterSeeder) Run() error {
	if err := s.seedCategories(); err != nil {
		return err
	}
	if err := s.seedUnits(); err != nil {
		return err
	}
	if err := s.seedWarehouses(); err != nil {
		return err
	}
	return s.seedPaymentMethods()
}

func (s *MasterSeeder) seedCategories() error {
	categories := []models.Category{
		{Name: "Makanan", Status: "active"},
		{Name: "Minuman", Status: "active"},
		{Name: "Elektronik", Status: "active"},
		{Name: "ATK", Status: "active"},
		{Name: "Sembako", Status: "active"},
	}
	for _, item := range categories {
		count, err := facades.Orm().Query().Model(&models.Category{}).Where("name", item.Name).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := facades.Orm().Query().Create(&item); err != nil {
			return err
		}
	}
	return nil
}

func (s *MasterSeeder) seedUnits() error {
	units := []models.Unit{
		{Name: "Pieces", Code: "pcs"},
		{Name: "Box", Code: "box"},
		{Name: "Pack", Code: "pack"},
		{Name: "Kilogram", Code: "kg"},
		{Name: "Liter", Code: "ltr"},
	}
	for _, item := range units {
		count, err := facades.Orm().Query().Model(&models.Unit{}).Where("code", item.Code).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := facades.Orm().Query().Create(&item); err != nil {
			return err
		}
	}
	return nil
}

func (s *MasterSeeder) seedWarehouses() error {
	warehouses := []models.Warehouse{
		{Code: "WH-001", Name: "Gudang Utama", Address: "Jl. Contoh No. 123", IsDefault: true, Status: "active"},
		{Code: "WH-002", Name: "Gudang Cabang 1", Address: "Jl. Cabang No. 1", IsDefault: false, Status: "active"},
	}
	for _, item := range warehouses {
		count, err := facades.Orm().Query().Model(&models.Warehouse{}).Where("code", item.Code).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := facades.Orm().Query().Create(&item); err != nil {
			return err
		}
	}
	return nil
}

func (s *MasterSeeder) seedPaymentMethods() error {
	methods := []models.PaymentMethod{
		{Code: "cash", Name: "Cash", IsActive: true},
		{Code: "qris", Name: "QRIS", IsActive: true},
		{Code: "debit", Name: "Debit Card", IsActive: true},
		{Code: "credit", Name: "Credit Card", IsActive: true},
		{Code: "transfer", Name: "Bank Transfer", IsActive: true},
	}
	for _, item := range methods {
		count, err := facades.Orm().Query().Model(&models.PaymentMethod{}).Where("code", item.Code).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := facades.Orm().Query().Create(&item); err != nil {
			return err
		}
	}
	return nil
}
