package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// WarehouseController — CRUD for warehouses (PRD §8.8).
type WarehouseController struct{}

func NewWarehouseController() *WarehouseController {
	return &WarehouseController{}
}

func (r *WarehouseController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.Warehouse{})
	if s := ctx.Request().Query("search"); s != "" {
		like := "%" + s + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	if status := ctx.Request().Query("status"); status != "" {
		query = query.Where("status", status)
	}

	var warehouses []models.Warehouse
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &warehouses, &total); err != nil {
		return serverError(ctx, "Failed to fetch warehouses")
	}
	return successMeta(ctx, "Warehouses retrieved successfully", warehouses, paginationMeta(page, perPage, total))
}

func (r *WarehouseController) Show(ctx http.Context) http.Response {
	var warehouse models.Warehouse
	if err := facades.Orm().Query().Find(&warehouse, ctx.Request().RouteInt("id")); err != nil || warehouse.ID == 0 {
		return notFound(ctx, "Warehouse not found")
	}
	return success(ctx, "Warehouse retrieved successfully", warehouse)
}

func (r *WarehouseController) Store(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"code": "required|max_len:20",
		"name": "required|max_len:100",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	code := ctx.Request().Input("code")
	if exists, _ := facades.Orm().Query().Model(&models.Warehouse{}).Where("code", code).Exists(); exists {
		return unprocessable(ctx, "A warehouse with this code already exists")
	}

	warehouse := models.Warehouse{
		Code:      code,
		Name:      ctx.Request().Input("name"),
		Address:   ctx.Request().Input("address"),
		Phone:     ctx.Request().Input("phone"),
		IsDefault: ctx.Request().Input("is_default") == "true",
		Status:    ctx.Request().Input("status", "active"),
	}
	if err := facades.Orm().Query().Create(&warehouse); err != nil {
		return serverError(ctx, "Failed to create warehouse")
	}
	return created(ctx, "Warehouse created successfully", warehouse)
}

func (r *WarehouseController) Update(ctx http.Context) http.Response {
	var warehouse models.Warehouse
	if err := facades.Orm().Query().Find(&warehouse, ctx.Request().RouteInt("id")); err != nil || warehouse.ID == 0 {
		return notFound(ctx, "Warehouse not found")
	}

	validator, err := ctx.Request().Validate(map[string]any{
		"code": "required|max_len:20",
		"name": "required|max_len:100",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	code := ctx.Request().Input("code", warehouse.Code)
	if exists, _ := facades.Orm().Query().Model(&models.Warehouse{}).
		Where("code", code).Where("id != ?", warehouse.ID).Exists(); exists {
		return unprocessable(ctx, "A warehouse with this code already exists")
	}

	warehouse.Code = code
	warehouse.Name = ctx.Request().Input("name", warehouse.Name)
	warehouse.Address = ctx.Request().Input("address", warehouse.Address)
	warehouse.Phone = ctx.Request().Input("phone", warehouse.Phone)
	warehouse.Status = ctx.Request().Input("status", warehouse.Status)
	if err := facades.Orm().Query().Save(&warehouse); err != nil {
		return serverError(ctx, "Failed to update warehouse")
	}
	return success(ctx, "Warehouse updated successfully", warehouse)
}

func (r *WarehouseController) Destroy(ctx http.Context) http.Response {
	var warehouse models.Warehouse
	if err := facades.Orm().Query().Find(&warehouse, ctx.Request().RouteInt("id")); err != nil || warehouse.ID == 0 {
		return notFound(ctx, "Warehouse not found")
	}
	if warehouse.IsDefault {
		return unprocessable(ctx, "The default warehouse cannot be deleted")
	}
	if _, err := facades.Orm().Query().Delete(&warehouse); err != nil {
		return serverError(ctx, "Failed to delete warehouse")
	}
	return success(ctx, "Warehouse deleted successfully", http.Json{})
}
