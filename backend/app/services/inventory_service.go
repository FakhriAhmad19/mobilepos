// Package services holds domain logic shared across controllers.
package services

import (
	"errors"

	"github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// ErrInsufficientStock is returned when a movement would drive stock negative
// (PRD §17 default business rule: stock >= 0).
var ErrInsufficientStock = errors.New("insufficient stock")

// MovementInput describes a single stock change to apply.
type MovementInput struct {
	ProductID     uint
	WarehouseID   uint
	Delta         int    // signed change applied to current stock
	Type          string // one of models.Movement*
	ReferenceType string
	ReferenceID   *uint
	UserID        *uint
	Notes         string
}

// ApplyMovement adjusts a product's stock in a warehouse and writes an
// immutable stock_movements ledger row (PRD §8.14). It MUST run inside a
// database transaction — pass the transaction's query as tx.
//
// It row-locks (or creates) the stock record, enforces the non-negative rule,
// and returns the quantities before and after the change.
func ApplyMovement(tx orm.Query, in MovementInput) (before, after int, err error) {
	var stock models.Stock
	// Find (not First) so a missing row yields an empty struct instead of an error.
	if err = tx.Where("product_id", in.ProductID).
		Where("warehouse_id", in.WarehouseID).
		LockForUpdate().Find(&stock); err != nil {
		return 0, 0, err
	}

	if stock.ID == 0 {
		// First-ever movement for this product/warehouse pair.
		stock = models.Stock{ProductID: in.ProductID, WarehouseID: in.WarehouseID}
		if err = tx.Create(&stock); err != nil {
			return 0, 0, err
		}
	}

	before = stock.Quantity
	after = before + in.Delta
	if after < 0 {
		return before, before, ErrInsufficientStock
	}

	stock.Quantity = after
	if err = tx.Save(&stock); err != nil {
		return before, before, err
	}

	movement := models.StockMovement{
		ProductID:      in.ProductID,
		WarehouseID:    in.WarehouseID,
		Type:           in.Type,
		Quantity:       in.Delta,
		QuantityBefore: before,
		QuantityAfter:  after,
		ReferenceType:  in.ReferenceType,
		ReferenceID:    in.ReferenceID,
		UserID:         in.UserID,
		Notes:          in.Notes,
	}
	if err = tx.Create(&movement); err != nil {
		return before, before, err
	}

	return before, after, nil
}
