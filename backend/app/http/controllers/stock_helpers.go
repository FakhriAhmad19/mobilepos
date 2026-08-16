package controllers

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// currentUserID returns the authenticated user's id, or nil if unavailable.
func currentUserID(ctx http.Context) *uint {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil || user.ID == 0 {
		return nil
	}
	id := user.ID
	return &id
}

// genRef builds a unique document reference like "SI-20260816-482915".
func genRef(prefix string) string {
	now := time.Now()
	return fmt.Sprintf("%s-%s-%06d", prefix, now.Format("20060102"), now.UnixNano()%1000000)
}

func productExists(id uint) bool {
	ok, _ := facades.Orm().Query().Model(&models.Product{}).Where("id", id).Exists()
	return ok
}

func warehouseExists(id uint) bool {
	ok, _ := facades.Orm().Query().Model(&models.Warehouse{}).Where("id", id).Exists()
	return ok
}

// validateWarehouseAndItems returns an error message, or "" if valid.
func validateWarehouseAndItems(warehouseID uint, itemCount int) string {
	if warehouseID == 0 {
		return "warehouse_id is required"
	}
	if !warehouseExists(warehouseID) {
		return "Selected warehouse does not exist"
	}
	if itemCount == 0 {
		return "At least one item is required"
	}
	return ""
}
