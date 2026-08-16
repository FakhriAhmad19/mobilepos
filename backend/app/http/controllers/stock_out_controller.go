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

// StockOutController — non-sale stock removal (PRD §8.11). Decrements stock and
// writes STOCK_OUT movements atomically; rejects removals that exceed stock.
type StockOutController struct{}

func NewStockOutController() *StockOutController {
	return &StockOutController{}
}

type stockOutItemReq struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type stockOutReq struct {
	WarehouseID uint              `json:"warehouse_id"`
	Reason      string            `json:"reason"`
	Notes       string            `json:"notes"`
	Items       []stockOutItemReq `json:"items"`
}

func (r *StockOutController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)
	query := facades.Orm().Query().Model(&models.StockOut{}).With("Warehouse").With("User")
	if wid := ctx.Request().QueryInt("warehouse_id", 0); wid > 0 {
		query = query.Where("warehouse_id", wid)
	}
	var records []models.StockOut
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &records, &total); err != nil {
		return serverError(ctx, "Failed to fetch stock-outs")
	}
	return successMeta(ctx, "Stock-outs retrieved successfully", records, paginationMeta(page, perPage, total))
}

func (r *StockOutController) Show(ctx http.Context) http.Response {
	var record models.StockOut
	if err := facades.Orm().Query().With("Warehouse").With("User").With("Items").With("Items.Product").
		Where("id", ctx.Request().RouteInt("id")).First(&record); err != nil || record.ID == 0 {
		return notFound(ctx, "Stock-out not found")
	}
	return success(ctx, "Stock-out retrieved successfully", record)
}

func (r *StockOutController) Store(ctx http.Context) http.Response {
	var req stockOutReq
	if err := ctx.Request().Bind(&req); err != nil {
		return serverError(ctx, "Invalid request body")
	}
	if msg := validateWarehouseAndItems(req.WarehouseID, len(req.Items)); msg != "" {
		return unprocessable(ctx, msg)
	}
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return unprocessable(ctx, "Each item quantity must be greater than zero")
		}
		if !productExists(item.ProductID) {
			return unprocessable(ctx, "One or more products do not exist")
		}
	}

	reason := req.Reason
	if reason == "" {
		reason = "manual"
	}
	userID := currentUserID(ctx)
	stockOut := models.StockOut{
		ReferenceNumber: genRef("SO"),
		WarehouseID:     req.WarehouseID,
		Reason:          reason,
		Notes:           req.Notes,
		Status:          "completed",
		Date:            carbon.NewDateTime(carbon.Now()),
	}
	if userID != nil {
		stockOut.UserID = *userID
	}

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Create(&stockOut); err != nil {
			return err
		}
		for _, item := range req.Items {
			line := models.StockOutItem{
				StockOutID: stockOut.ID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
			}
			if err := tx.Create(&line); err != nil {
				return err
			}
			refID := stockOut.ID
			if _, _, err := services.ApplyMovement(tx, services.MovementInput{
				ProductID:     item.ProductID,
				WarehouseID:   req.WarehouseID,
				Delta:         -item.Quantity,
				Type:          models.MovementStockOut,
				ReferenceType: "stock_out",
				ReferenceID:   &refID,
				UserID:        userID,
				Notes:         req.Notes,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if errors.Is(err, services.ErrInsufficientStock) {
		return unprocessable(ctx, "Insufficient stock for one or more items")
	}
	if err != nil {
		return serverError(ctx, "Failed to record stock-out")
	}

	_ = facades.Orm().Query().With("Items").With("Warehouse").Where("id", stockOut.ID).First(&stockOut)
	return created(ctx, "Stock-out recorded successfully", stockOut)
}
