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

// OrderController — POS checkout, transaction history and receipts
// (PRD §8.19–§8.22).
type OrderController struct{}

func NewOrderController() *OrderController {
	return &OrderController{}
}

type checkoutItemReq struct {
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Discount  float64 `json:"discount"`
}

type checkoutPaymentReq struct {
	PaymentMethodID uint    `json:"payment_method_id"`
	AmountPaid      float64 `json:"amount_paid"`
	Reference       string  `json:"reference"`
}

type checkoutReq struct {
	WarehouseID uint                 `json:"warehouse_id"`
	CustomerID  *uint                `json:"customer_id"`
	Discount    float64              `json:"discount"`
	Tax         float64              `json:"tax"`
	Notes       string               `json:"notes"`
	Items       []checkoutItemReq    `json:"items"`
	Payments    []checkoutPaymentReq `json:"payments"`
}

// Store performs an atomic checkout (PRD §8.19, §18): it creates the order,
// its line items and payments, and reduces stock via SALE movements — all in a
// single transaction that rolls back on any failure.
func (r *OrderController) Store(ctx http.Context) http.Response {
	var req checkoutReq
	if err := ctx.Request().Bind(&req); err != nil {
		return serverError(ctx, "Invalid request body")
	}
	if req.WarehouseID == 0 || !warehouseExists(req.WarehouseID) {
		return unprocessable(ctx, "Selected warehouse does not exist")
	}
	if len(req.Items) == 0 {
		return unprocessable(ctx, "Cart is empty")
	}
	if len(req.Payments) == 0 {
		return unprocessable(ctx, "At least one payment is required")
	}

	// Build line items with server-authoritative prices.
	var items []models.OrderItem
	var subtotal float64
	for _, it := range req.Items {
		if it.Quantity <= 0 {
			return unprocessable(ctx, "Each item quantity must be greater than zero")
		}
		var product models.Product
		if err := facades.Orm().Query().Where("id", it.ProductID).First(&product); err != nil || product.ID == 0 {
			return unprocessable(ctx, "One or more products do not exist")
		}
		lineSubtotal := services.LineSubtotal(it.Quantity, product.SellingPrice, it.Discount)
		subtotal += lineSubtotal
		items = append(items, models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			SKU:         product.SKU,
			UnitPrice:   product.SellingPrice,
			Quantity:    it.Quantity,
			Discount:    it.Discount,
			Subtotal:    lineSubtotal,
		})
	}

	grandTotal := services.GrandTotal(subtotal, req.Discount, req.Tax)

	// Validate and total the payments.
	var totalPaid float64
	for _, p := range req.Payments {
		if p.PaymentMethodID == 0 {
			return unprocessable(ctx, "Each payment requires a payment_method_id")
		}
		if ok, _ := facades.Orm().Query().Model(&models.PaymentMethod{}).Where("id", p.PaymentMethodID).Exists(); !ok {
			return unprocessable(ctx, "Invalid payment method")
		}
		totalPaid += p.AmountPaid
	}
	if !services.PaymentCovers(totalPaid, grandTotal) {
		return unprocessable(ctx, "Insufficient payment amount")
	}
	change := totalPaid - grandTotal

	userID := currentUserID(ctx)
	order := models.Order{
		OrderNumber: genRef("TRX"),
		CustomerID:  req.CustomerID,
		WarehouseID: req.WarehouseID,
		Subtotal:    subtotal,
		Discount:    req.Discount,
		Tax:         req.Tax,
		GrandTotal:  grandTotal,
		Status:      "completed",
		Notes:       req.Notes,
		OrderedAt:   carbon.NewDateTime(carbon.Now()),
	}
	if userID != nil {
		order.UserID = *userID
	}

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		if err := tx.Create(&order); err != nil {
			return err
		}

		// Line items + stock reduction.
		for i := range items {
			items[i].OrderID = order.ID
			if err := tx.Create(&items[i]); err != nil {
				return err
			}
			refID := order.ID
			if _, _, err := services.ApplyMovement(tx, services.MovementInput{
				ProductID:     items[i].ProductID,
				WarehouseID:   req.WarehouseID,
				Delta:         -items[i].Quantity,
				Type:          models.MovementSale,
				ReferenceType: "order",
				ReferenceID:   &refID,
				UserID:        userID,
				Notes:         order.OrderNumber,
			}); err != nil {
				return err
			}
		}

		// Payments — allocate the amount due across methods; change on the last.
		remaining := grandTotal
		for idx, p := range req.Payments {
			alloc := p.AmountPaid
			if alloc > remaining {
				alloc = remaining
			}
			remaining -= alloc
			payment := models.Payment{
				OrderID:         order.ID,
				PaymentMethodID: p.PaymentMethodID,
				Amount:          alloc,
				AmountPaid:      p.AmountPaid,
				Reference:       p.Reference,
				Status:          "paid",
				PaidAt:          carbon.NewDateTime(carbon.Now()),
			}
			if idx == len(req.Payments)-1 {
				payment.Change = change
			}
			if err := tx.Create(&payment); err != nil {
				return err
			}
		}
		return nil
	})
	if errors.Is(err, services.ErrInsufficientStock) {
		return unprocessable(ctx, "Insufficient stock for one or more items")
	}
	if err != nil {
		return serverError(ctx, "Checkout failed")
	}

	receipt, _ := r.find(int(order.ID))
	return created(ctx, "Order created successfully", receipt)
}

func (r *OrderController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.Order{}).With("User").With("Customer").With("Warehouse")
	if uid := ctx.Request().QueryInt("user_id", 0); uid > 0 {
		query = query.Where("user_id", uid)
	}
	if cid := ctx.Request().QueryInt("customer_id", 0); cid > 0 {
		query = query.Where("customer_id", cid)
	}
	if status := ctx.Request().Query("status"); status != "" {
		query = query.Where("status", status)
	}
	if from := ctx.Request().Query("date_from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := ctx.Request().Query("date_to"); to != "" {
		query = query.Where("created_at <= ?", to+" 23:59:59")
	}
	if s := ctx.Request().Query("search"); s != "" {
		query = query.Where("order_number LIKE ?", "%"+s+"%")
	}

	var orders []models.Order
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &orders, &total); err != nil {
		return serverError(ctx, "Failed to fetch orders")
	}
	return successMeta(ctx, "Orders retrieved successfully", orders, paginationMeta(page, perPage, total))
}

func (r *OrderController) Show(ctx http.Context) http.Response {
	order, ok := r.find(ctx.Request().RouteInt("id"))
	if !ok {
		return notFound(ctx, "Order not found")
	}
	return success(ctx, "Order retrieved successfully", order)
}

// Void cancels a completed order and returns its items to stock via RETURN
// movements (PRD §8.14). Admin only.
// POST /orders/{id}/void
func (r *OrderController) Void(ctx http.Context) http.Response {
	var order models.Order
	if err := facades.Orm().Query().With("OrderItems").
		Where("id", ctx.Request().RouteInt("id")).First(&order); err != nil || order.ID == 0 {
		return notFound(ctx, "Order not found")
	}
	if order.Status == "cancelled" {
		return unprocessable(ctx, "This order is already cancelled")
	}

	userID := currentUserID(ctx)
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		for _, item := range order.OrderItems {
			refID := order.ID
			if _, _, err := services.ApplyMovement(tx, services.MovementInput{
				ProductID:     item.ProductID,
				WarehouseID:   order.WarehouseID,
				Delta:         item.Quantity,
				Type:          models.MovementReturn,
				ReferenceType: "order_void",
				ReferenceID:   &refID,
				UserID:        userID,
				Notes:         "Void " + order.OrderNumber,
			}); err != nil {
				return err
			}
		}
		order.Status = "cancelled"
		return tx.Save(&order)
	})
	if err != nil {
		return serverError(ctx, "Failed to void order")
	}

	receipt, _ := r.find(int(order.ID))
	return success(ctx, "Order voided successfully", receipt)
}

func (r *OrderController) find(id int) (models.Order, bool) {
	var order models.Order
	if err := facades.Orm().Query().
		With("User").With("Customer").With("Warehouse").
		With("OrderItems").With("Payments").With("Payments.PaymentMethod").
		Where("id", id).First(&order); err != nil || order.ID == 0 {
		return order, false
	}
	return order, true
}
