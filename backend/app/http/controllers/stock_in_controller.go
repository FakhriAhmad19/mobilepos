package controllers

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"

	"goravel/app/models"
	"goravel/app/services"
)

// StockInController — goods receiving (PRD §8.10). Each receipt increments
// stock and writes STOCK_IN movements, all in one transaction (PRD §18).
type StockInController struct{}

func NewStockInController() *StockInController {
	return &StockInController{}
}

type stockInItemReq struct {
	ProductID     uint    `json:"product_id"`
	Quantity      int     `json:"quantity"`
	PurchasePrice float64 `json:"purchase_price"`
}

type stockInReq struct {
	SupplierID  *uint            `json:"supplier_id"`
	WarehouseID uint             `json:"warehouse_id"`
	Notes       string           `json:"notes"`
	Items       []stockInItemReq `json:"items"`
}

func (r *StockInController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)
	query := facades.Orm().Query().Model(&models.StockIn{}).With("Supplier").With("Warehouse").With("User")
	if wid := ctx.Request().QueryInt("warehouse_id", 0); wid > 0 {
		query = query.Where("warehouse_id", wid)
	}
	var records []models.StockIn
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &records, &total); err != nil {
		return serverError(ctx, "Failed to fetch stock-ins")
	}
	return successMeta(ctx, "Stock-ins retrieved successfully", records, paginationMeta(page, perPage, total))
}

func (r *StockInController) Show(ctx http.Context) http.Response {
	var record models.StockIn
	if err := facades.Orm().Query().With("Supplier").With("Warehouse").With("User").
		With("Items").With("Items.Product").
		Where("id", ctx.Request().RouteInt("id")).First(&record); err != nil || record.ID == 0 {
		return notFound(ctx, "Stock-in not found")
	}
	return success(ctx, "Stock-in retrieved successfully", record)
}

func (r *StockInController) Store(ctx http.Context) http.Response {
	var req stockInReq
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

	userID := currentUserID(ctx)
	stockIn := models.StockIn{
		ReferenceNumber: genRef("SI"),
		SupplierID:      req.SupplierID,
		WarehouseID:     req.WarehouseID,
		Notes:           req.Notes,
		Status:          "completed",
		Date:            carbon.NewDateTime(carbon.Now()),
	}
	if userID != nil {
		stockIn.UserID = *userID
	}

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Create(&stockIn); err != nil {
			return err
		}
		for _, item := range req.Items {
			line := models.StockInItem{
				StockInID:     stockIn.ID,
				ProductID:     item.ProductID,
				Quantity:      item.Quantity,
				PurchasePrice: item.PurchasePrice,
				Subtotal:      float64(item.Quantity) * item.PurchasePrice,
			}
			if err := tx.Create(&line); err != nil {
				return err
			}
			refID := stockIn.ID
			if _, _, err := services.ApplyMovement(tx, services.MovementInput{
				ProductID:     item.ProductID,
				WarehouseID:   req.WarehouseID,
				Delta:         item.Quantity,
				Type:          models.MovementStockIn,
				ReferenceType: "stock_in",
				ReferenceID:   &refID,
				UserID:        userID,
				Notes:         req.Notes,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return serverError(ctx, "Failed to record stock-in")
	}

	_ = facades.Orm().Query().With("Items").With("Warehouse").With("Supplier").
		Where("id", stockIn.ID).First(&stockIn)
	return created(ctx, "Stock-in recorded successfully", stockIn)
}
