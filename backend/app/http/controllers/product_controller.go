package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// ProductController — CRUD for products (PRD §8.4, §8.16 search).
type ProductController struct{}

func NewProductController() *ProductController {
	return &ProductController{}
}

type productRequest struct {
	SKU           string  `json:"sku"`
	Barcode       string  `json:"barcode"`
	Name          string  `json:"name"`
	CategoryID    uint    `json:"category_id"`
	UnitID        uint    `json:"unit_id"`
	PurchasePrice float64 `json:"purchase_price"`
	SellingPrice  float64 `json:"selling_price"`
	MinimumStock  int     `json:"minimum_stock"`
	Image         string  `json:"image"`
	Description   string  `json:"description"`
	Status        string  `json:"status"`
}

func (r *ProductController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.Product{}).With("Category").With("Unit")
	if s := ctx.Request().Query("search"); s != "" {
		like := "%" + s + "%"
		query = query.Where("name LIKE ? OR sku LIKE ? OR barcode LIKE ?", like, like, like)
	}
	if categoryID := ctx.Request().QueryInt("category_id", 0); categoryID > 0 {
		query = query.Where("category_id", categoryID)
	}
	if status := ctx.Request().Query("status"); status != "" {
		query = query.Where("status", status)
	}

	var products []models.Product
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &products, &total); err != nil {
		return serverError(ctx, "Failed to fetch products")
	}

	return successMeta(ctx, "Products retrieved successfully", products, paginationMeta(page, perPage, total))
}

func (r *ProductController) Show(ctx http.Context) http.Response {
	var product models.Product
	if err := facades.Orm().Query().With("Category").With("Unit").
		Where("id", ctx.Request().RouteInt("id")).First(&product); err != nil || product.ID == 0 {
		return notFound(ctx, "Product not found")
	}
	return success(ctx, "Product retrieved successfully", product)
}

// Barcode looks up a single product by its barcode for the POS scanner
// (PRD §8.17). Returns 404 ("Product Not Found") when no match exists.
// GET /pos/scan/{barcode}
func (r *ProductController) Barcode(ctx http.Context) http.Response {
	code := ctx.Request().Route("barcode")
	var product models.Product
	if err := facades.Orm().Query().With("Category").With("Unit").
		Where("barcode", code).First(&product); err != nil || product.ID == 0 {
		return notFound(ctx, "Product not found")
	}
	return success(ctx, "Product retrieved successfully", product)
}

func (r *ProductController) Store(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"sku":           "required|max_len:50",
		"name":          "required|max_len:150",
		"category_id":   "required",
		"unit_id":       "required",
		"selling_price": "required",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	var req productRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return serverError(ctx, "Invalid request body")
	}

	if msg := r.validateRefs(req.CategoryID, req.UnitID); msg != "" {
		return unprocessable(ctx, msg)
	}
	if exists, _ := facades.Orm().Query().Model(&models.Product{}).Where("sku", req.SKU).Exists(); exists {
		return unprocessable(ctx, "A product with this SKU already exists")
	}

	product := models.Product{
		SKU:           req.SKU,
		Barcode:       req.Barcode,
		Name:          req.Name,
		CategoryID:    req.CategoryID,
		UnitID:        req.UnitID,
		PurchasePrice: req.PurchasePrice,
		SellingPrice:  req.SellingPrice,
		MinimumStock:  req.MinimumStock,
		Image:         req.Image,
		Description:   req.Description,
		Status:        defaultStr(req.Status, "active"),
	}
	if err := facades.Orm().Query().Create(&product); err != nil {
		return serverError(ctx, "Failed to create product")
	}
	_ = facades.Orm().Query().With("Category").With("Unit").Where("id", product.ID).First(&product)
	return created(ctx, "Product created successfully", product)
}

func (r *ProductController) Update(ctx http.Context) http.Response {
	var product models.Product
	if err := facades.Orm().Query().Find(&product, ctx.Request().RouteInt("id")); err != nil || product.ID == 0 {
		return notFound(ctx, "Product not found")
	}

	validator, err := ctx.Request().Validate(map[string]any{
		"sku":           "required|max_len:50",
		"name":          "required|max_len:150",
		"category_id":   "required",
		"unit_id":       "required",
		"selling_price": "required",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	var req productRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return serverError(ctx, "Invalid request body")
	}

	if msg := r.validateRefs(req.CategoryID, req.UnitID); msg != "" {
		return unprocessable(ctx, msg)
	}
	if exists, _ := facades.Orm().Query().Model(&models.Product{}).
		Where("sku", req.SKU).Where("id != ?", product.ID).Exists(); exists {
		return unprocessable(ctx, "A product with this SKU already exists")
	}

	product.SKU = req.SKU
	product.Barcode = req.Barcode
	product.Name = req.Name
	product.CategoryID = req.CategoryID
	product.UnitID = req.UnitID
	product.PurchasePrice = req.PurchasePrice
	product.SellingPrice = req.SellingPrice
	product.MinimumStock = req.MinimumStock
	product.Image = req.Image
	product.Description = req.Description
	if req.Status != "" {
		product.Status = req.Status
	}
	if err := facades.Orm().Query().Save(&product); err != nil {
		return serverError(ctx, "Failed to update product")
	}
	_ = facades.Orm().Query().With("Category").With("Unit").Where("id", product.ID).First(&product)
	return success(ctx, "Product updated successfully", product)
}

func (r *ProductController) Destroy(ctx http.Context) http.Response {
	var product models.Product
	if err := facades.Orm().Query().Find(&product, ctx.Request().RouteInt("id")); err != nil || product.ID == 0 {
		return notFound(ctx, "Product not found")
	}
	if _, err := facades.Orm().Query().Delete(&product); err != nil {
		return serverError(ctx, "Failed to delete product")
	}
	return success(ctx, "Product deleted successfully", http.Json{})
}

// validateRefs ensures the referenced category and unit exist.
func (r *ProductController) validateRefs(categoryID, unitID uint) string {
	if ok, _ := facades.Orm().Query().Model(&models.Category{}).Where("id", categoryID).Exists(); !ok {
		return "Selected category does not exist"
	}
	if ok, _ := facades.Orm().Query().Model(&models.Unit{}).Where("id", unitID).Exists(); !ok {
		return "Selected unit does not exist"
	}
	return ""
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
