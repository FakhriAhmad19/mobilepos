package seeders

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// ProductSeeder seeds sample products with initial stock in the default
// warehouse (PRD §8.4, §8.9).
type ProductSeeder struct{}

func (s *ProductSeeder) Signature() string {
	return "ProductSeeder"
}

type productSeed struct {
	SKU           string
	Barcode       string
	Name          string
	Category      string
	Unit          string
	PurchasePrice float64
	SellingPrice  float64
	MinimumStock  int
	InitialStock  int
}

func (s *ProductSeeder) Run() error {
	// Resolve lookup maps.
	categoryIDs, err := s.lookupCategories()
	if err != nil {
		return err
	}
	unitIDs, err := s.lookupUnits()
	if err != nil {
		return err
	}

	var defaultWarehouse models.Warehouse
	if err := facades.Orm().Query().Where("is_default", true).First(&defaultWarehouse); err != nil {
		return err
	}

	seeds := []productSeed{
		{"SKU-001", "8991234500011", "Air Mineral 600ml", "Minuman", "pcs", 2000, 3500, 24, 100},
		{"SKU-002", "8991234500028", "Kopi Sachet", "Minuman", "pcs", 1000, 1500, 40, 200},
		{"SKU-003", "8991234500035", "Mie Instan Goreng", "Makanan", "pcs", 2500, 3500, 30, 150},
		{"SKU-004", "8991234500042", "Beras 5kg", "Sembako", "pack", 55000, 65000, 10, 40},
		{"SKU-005", "8991234500059", "Buku Tulis 38 lembar", "ATK", "pcs", 3000, 5000, 20, 80},
		{"SKU-006", "8991234500066", "Pulpen Hitam", "ATK", "pcs", 1500, 2500, 25, 120},
	}

	for _, seed := range seeds {
		count, err := facades.Orm().Query().Model(&models.Product{}).Where("sku", seed.SKU).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		product := models.Product{
			SKU:           seed.SKU,
			Barcode:       seed.Barcode,
			Name:          seed.Name,
			CategoryID:    categoryIDs[seed.Category],
			UnitID:        unitIDs[seed.Unit],
			PurchasePrice: seed.PurchasePrice,
			SellingPrice:  seed.SellingPrice,
			MinimumStock:  seed.MinimumStock,
			Status:        "active",
		}
		if err := facades.Orm().Query().Create(&product); err != nil {
			return err
		}

		stock := models.Stock{
			ProductID:   product.ID,
			WarehouseID: defaultWarehouse.ID,
			Quantity:    seed.InitialStock,
			Reserved:    0,
		}
		if err := facades.Orm().Query().Create(&stock); err != nil {
			return err
		}
	}

	return nil
}

func (s *ProductSeeder) lookupCategories() (map[string]uint, error) {
	var categories []models.Category
	if err := facades.Orm().Query().Find(&categories); err != nil {
		return nil, err
	}
	out := map[string]uint{}
	for _, c := range categories {
		out[c.Name] = c.ID
	}
	return out, nil
}

func (s *ProductSeeder) lookupUnits() (map[string]uint, error) {
	var units []models.Unit
	if err := facades.Orm().Query().Find(&units); err != nil {
		return nil, err
	}
	out := map[string]uint{}
	for _, u := range units {
		out[u.Code] = u.ID
	}
	return out, nil
}
