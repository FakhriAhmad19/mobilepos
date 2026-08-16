package controllers

import (
	"errors"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"

	"goravel/app/models"
	"goravel/app/services"
)

// StockOpnameController — physical stock count (PRD §8.13). Creating an opname
// snapshots system stock; completing it applies the differences as adjustments.
type StockOpnameController struct{}

func NewStockOpnameController() *StockOpnameController {
	return &StockOpnameController{}
}

type opnameItemReq struct {
	ProductID     uint `json:"product_id"`
	PhysicalStock int  `json:"physical_stock"`
}

type opnameReq struct {
	WarehouseID uint            `json:"warehouse_id"`
	Notes       string          `json:"notes"`
	Items       []opnameItemReq `json:"items"`
}

func (r *StockOpnameController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)
	query := facades.Orm().Query().Model(&models.StockOpname{}).With("Warehouse").With("User")
	if wid := ctx.Request().QueryInt("warehouse_id", 0); wid > 0 {
		query = query.Where("warehouse_id", wid)
	}
	if status := ctx.Request().Query("status"); status != "" {
		query = query.Where("status", status)
	}
	var records []models.StockOpname
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &records, &total); err != nil {
		return serverError(ctx, "Failed to fetch stock opnames")
	}
	return successMeta(ctx, "Stock opnames retrieved successfully", records, paginationMeta(page, perPage, total))
}

func (r *StockOpnameController) Show(ctx http.Context) http.Response {
	record, ok := r.find(ctx.Request().RouteInt("id"))
	if !ok {
		return notFound(ctx, "Stock opname not found")
	}
	return success(ctx, "Stock opname retrieved successfully", record)
}

func (r *StockOpnameController) Store(ctx http.Context) http.Response {
	var req opnameReq
	if err := ctx.Request().Bind(&req); err != nil {
		return serverError(ctx, "Invalid request body")
	}
	if msg := validateWarehouseAndItems(req.WarehouseID, len(req.Items)); msg != "" {
		return unprocessable(ctx, msg)
	}
	for _, item := range req.Items {
		if !productExists(item.ProductID) {
			return unprocessable(ctx, "One or more products do not exist")
		}
	}

	userID := currentUserID(ctx)
	opname := models.StockOpname{
		ReferenceNumber: genRef("OPN"),
		WarehouseID:     req.WarehouseID,
		Status:          "draft",
		Notes:           req.Notes,
		StartedAt:       carbon.NewDateTime(carbon.Now()),
	}
	if userID != nil {
		opname.UserID = *userID
	}

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Create(&opname); err != nil {
			return err
		}
		for _, item := range req.Items {
			system := currentQuantity(tx, item.ProductID, req.WarehouseID)
			line := models.StockOpnameItem{
				StockOpnameID: opname.ID,
				ProductID:     item.ProductID,
				SystemStock:   system,
				PhysicalStock: item.PhysicalStock,
				Difference:    item.PhysicalStock - system,
			}
			if err := tx.Create(&line); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return serverError(ctx, "Failed to create stock opname")
	}

	record, _ := r.find(int(opname.ID))
	return created(ctx, "Stock opname created successfully", record)
}

// Complete applies each counted difference to stock as an ADJUSTMENT movement
// and marks the opname completed (PRD §8.13).
// POST /stock-opnames/{id}/complete
func (r *StockOpnameController) Complete(ctx http.Context) http.Response {
	var opname models.StockOpname
	if err := facades.Orm().Query().With("Items").
		Where("id", ctx.Request().RouteInt("id")).First(&opname); err != nil || opname.ID == 0 {
		return notFound(ctx, "Stock opname not found")
	}
	if opname.Status == "completed" {
		return unprocessable(ctx, "This stock opname is already completed")
	}

	userID := currentUserID(ctx)
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		for _, item := range opname.Items {
			current := currentQuantity(tx, item.ProductID, opname.WarehouseID)
			delta := item.PhysicalStock - current
			if delta != 0 {
				refID := opname.ID
				if _, _, err := services.ApplyMovement(tx, services.MovementInput{
					ProductID:     item.ProductID,
					WarehouseID:   opname.WarehouseID,
					Delta:         delta,
					Type:          models.MovementAdjustment,
					ReferenceType: "stock_opname",
					ReferenceID:   &refID,
					UserID:        userID,
					Notes:         "Stock opname " + opname.ReferenceNumber,
				}); err != nil {
					return err
				}
			}
		}
		opname.Status = "completed"
		opname.CompletedAt = carbon.NewDateTime(carbon.Now())
		return tx.Save(&opname)
	})
	if errors.Is(err, services.ErrInsufficientStock) {
		return unprocessable(ctx, "A counted value would drive stock negative")
	}
	if err != nil {
		return serverError(ctx, "Failed to complete stock opname")
	}

	record, _ := r.find(int(opname.ID))
	return success(ctx, "Stock opname completed successfully", record)
}

func (r *StockOpnameController) find(id int) (models.StockOpname, bool) {
	var record models.StockOpname
	if err := facades.Orm().Query().With("Warehouse").With("User").
		With("Items").With("Items.Product").
		Where("id", id).First(&record); err != nil || record.ID == 0 {
		return record, false
	}
	return record, true
}

// currentQuantity reads the current stock quantity (0 if no row yet).
func currentQuantity(tx orm.Query, productID, warehouseID uint) int {
	var stock models.Stock
	_ = tx.Where("product_id", productID).Where("warehouse_id", warehouseID).Find(&stock)
	return stock.Quantity
}
