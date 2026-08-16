package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// SupplierController — CRUD for suppliers (PRD §8.6).
type SupplierController struct{}

func NewSupplierController() *SupplierController {
	return &SupplierController{}
}

func (r *SupplierController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.Supplier{})
	if s := ctx.Request().Query("search"); s != "" {
		like := "%" + s + "%"
		query = query.Where("name LIKE ? OR contact_person LIKE ? OR phone LIKE ?", like, like, like)
	}
	if status := ctx.Request().Query("status"); status != "" {
		query = query.Where("status", status)
	}

	var suppliers []models.Supplier
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &suppliers, &total); err != nil {
		return serverError(ctx, "Failed to fetch suppliers")
	}
	return successMeta(ctx, "Suppliers retrieved successfully", suppliers, paginationMeta(page, perPage, total))
}

func (r *SupplierController) Show(ctx http.Context) http.Response {
	var supplier models.Supplier
	if err := facades.Orm().Query().Find(&supplier, ctx.Request().RouteInt("id")); err != nil || supplier.ID == 0 {
		return notFound(ctx, "Supplier not found")
	}
	return success(ctx, "Supplier retrieved successfully", supplier)
}

func (r *SupplierController) Store(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"name":  "required|max_len:150",
		"email": "email",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	supplier := models.Supplier{
		Name:          ctx.Request().Input("name"),
		ContactPerson: ctx.Request().Input("contact_person"),
		Phone:         ctx.Request().Input("phone"),
		Email:         ctx.Request().Input("email"),
		Address:       ctx.Request().Input("address"),
		Status:        ctx.Request().Input("status", "active"),
	}
	if err := facades.Orm().Query().Create(&supplier); err != nil {
		return serverError(ctx, "Failed to create supplier")
	}
	return created(ctx, "Supplier created successfully", supplier)
}

func (r *SupplierController) Update(ctx http.Context) http.Response {
	var supplier models.Supplier
	if err := facades.Orm().Query().Find(&supplier, ctx.Request().RouteInt("id")); err != nil || supplier.ID == 0 {
		return notFound(ctx, "Supplier not found")
	}

	validator, err := ctx.Request().Validate(map[string]any{
		"name":  "required|max_len:150",
		"email": "email",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	supplier.Name = ctx.Request().Input("name", supplier.Name)
	supplier.ContactPerson = ctx.Request().Input("contact_person", supplier.ContactPerson)
	supplier.Phone = ctx.Request().Input("phone", supplier.Phone)
	supplier.Email = ctx.Request().Input("email", supplier.Email)
	supplier.Address = ctx.Request().Input("address", supplier.Address)
	supplier.Status = ctx.Request().Input("status", supplier.Status)
	if err := facades.Orm().Query().Save(&supplier); err != nil {
		return serverError(ctx, "Failed to update supplier")
	}
	return success(ctx, "Supplier updated successfully", supplier)
}

func (r *SupplierController) Destroy(ctx http.Context) http.Response {
	var supplier models.Supplier
	if err := facades.Orm().Query().Find(&supplier, ctx.Request().RouteInt("id")); err != nil || supplier.ID == 0 {
		return notFound(ctx, "Supplier not found")
	}
	if _, err := facades.Orm().Query().Delete(&supplier); err != nil {
		return serverError(ctx, "Failed to delete supplier")
	}
	return success(ctx, "Supplier deleted successfully", http.Json{})
}
