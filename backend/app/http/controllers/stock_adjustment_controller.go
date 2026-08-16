package controllers

import (
	"errors"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/app/services"
)

// StockAdjustmentController — stock correction to a counted value (PRD §8.12).
type StockAdjustmentController struct{}

func NewStockAdjustmentController() *StockAdjustmentController {
	return &StockAdjustmentController{}
}

type stockAdjustmentReq struct {
	WarehouseID   uint   `json:"warehouse_id"`
	ProductID     uint   `json:"product_id"`
	PhysicalStock int    `json:"physical_stock"`
	Reason        string `json:"reason"`
	Notes         string `json:"notes"`
}

func (r *StockAdjustmentController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)
	query := facades.Orm().Query().Model(&models.StockAdjustment{}).With("Warehouse").With("Product").With("User")
	if wid := ctx.Request().QueryInt("warehouse_id", 0); wid > 0 {
		query = query.Where("warehouse_id", wid)
	}
	if pid := ctx.Request().QueryInt("product_id", 0); pid > 0 {
		query = query.Where("product_id", pid)
	}
	var records []models.StockAdjustment
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &records, &total); err != nil {
		return serverError(ctx, "Failed to fetch adjustments")
	}
	return successMeta(ctx, "Adjustments retrieved successfully", records, paginationMeta(page, perPage, total))
}

func (r *StockAdjustmentController) Show(ctx http.Context) http.Response {
	var record models.StockAdjustment
	if err := facades.Orm().Query().With("Warehouse").With("Product").With("User").
		Where("id", ctx.Request().RouteInt("id")).First(&record); err != nil || record.ID == 0 {
		return notFound(ctx, "Adjustment not found")
	}
	return success(ctx, "Adjustment retrieved successfully", record)
}

func (r *StockAdjustmentController) Store(ctx http.Context) http.Response {
	var req stockAdjustmentReq
	if err := ctx.Request().Bind(&req); err != nil {
		return serverError(ctx, "Invalid request body")
	}
	if req.WarehouseID == 0 || !warehouseExists(req.WarehouseID) {
		return unprocessable(ctx, "Selected warehouse does not exist")
	}
	if req.ProductID == 0 || !productExists(req.ProductID) {
		return unprocessable(ctx, "Selected product does not exist")
	}
	if req.PhysicalStock < 0 {
		return unprocessable(ctx, "Physical stock cannot be negative")
	}

	userID := currentUserID(ctx)
	adjustment := models.StockAdjustment{
		ReferenceNumber: genRef("ADJ"),
		WarehouseID:     req.WarehouseID,
		ProductID:       req.ProductID,
		Reason:          req.Reason,
		Notes:           req.Notes,
	}
	if userID != nil {
		adjustment.UserID = *userID
	}

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		var stock models.Stock
		if err := tx.Where("product_id", req.ProductID).Where("warehouse_id", req.WarehouseID).
			LockForUpdate().Find(&stock); err != nil {
			return err
		}
		before := stock.Quantity
		delta := req.PhysicalStock - before

		before2, after, err := services.ApplyMovement(tx, services.MovementInput{
			ProductID:     req.ProductID,
			WarehouseID:   req.WarehouseID,
			Delta:         delta,
			Type:          models.MovementAdjustment,
			ReferenceType: "stock_adjustment",
			UserID:        userID,
			Notes:         req.Notes,
		})
		if err != nil {
			return err
		}

		adjustment.QuantityBefore = before2
		adjustment.QuantityAfter = after
		adjustment.Difference = delta
		return tx.Create(&adjustment)
	})
	if errors.Is(err, services.ErrInsufficientStock) {
		return unprocessable(ctx, "Adjustment would drive stock negative")
	}
	if err != nil {
		return serverError(ctx, "Failed to record adjustment")
	}

	return created(ctx, "Adjustment recorded successfully", adjustment)
}
