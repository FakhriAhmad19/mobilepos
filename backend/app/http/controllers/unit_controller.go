package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// UnitController — CRUD for units of measure (PRD §7).
type UnitController struct{}

func NewUnitController() *UnitController {
	return &UnitController{}
}

func (r *UnitController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.Unit{})
	if s := ctx.Request().Query("search"); s != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+s+"%", "%"+s+"%")
	}

	var units []models.Unit
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &units, &total); err != nil {
		return serverError(ctx, "Failed to fetch units")
	}

	return successMeta(ctx, "Units retrieved successfully", units, paginationMeta(page, perPage, total))
}

func (r *UnitController) Show(ctx http.Context) http.Response {
	var unit models.Unit
	if err := facades.Orm().Query().Find(&unit, ctx.Request().RouteInt("id")); err != nil || unit.ID == 0 {
		return notFound(ctx, "Unit not found")
	}
	return success(ctx, "Unit retrieved successfully", unit)
}

func (r *UnitController) Store(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"name": "required|max_len:50",
		"code": "required|max_len:20",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	code := ctx.Request().Input("code")
	if exists, _ := facades.Orm().Query().Model(&models.Unit{}).Where("code", code).Exists(); exists {
		return unprocessable(ctx, "A unit with this code already exists")
	}

	unit := models.Unit{
		Name:        ctx.Request().Input("name"),
		Code:        code,
		Description: ctx.Request().Input("description"),
	}
	if err := facades.Orm().Query().Create(&unit); err != nil {
		return serverError(ctx, "Failed to create unit")
	}
	return created(ctx, "Unit created successfully", unit)
}

func (r *UnitController) Update(ctx http.Context) http.Response {
	var unit models.Unit
	if err := facades.Orm().Query().Find(&unit, ctx.Request().RouteInt("id")); err != nil || unit.ID == 0 {
		return notFound(ctx, "Unit not found")
	}

	validator, err := ctx.Request().Validate(map[string]any{
		"name": "required|max_len:50",
		"code": "required|max_len:20",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	code := ctx.Request().Input("code", unit.Code)
	if exists, _ := facades.Orm().Query().Model(&models.Unit{}).Where("code", code).Where("id != ?", unit.ID).Exists(); exists {
		return unprocessable(ctx, "A unit with this code already exists")
	}

	unit.Name = ctx.Request().Input("name", unit.Name)
	unit.Code = code
	unit.Description = ctx.Request().Input("description", unit.Description)
	if err := facades.Orm().Query().Save(&unit); err != nil {
		return serverError(ctx, "Failed to update unit")
	}
	return success(ctx, "Unit updated successfully", unit)
}

func (r *UnitController) Destroy(ctx http.Context) http.Response {
	var unit models.Unit
	if err := facades.Orm().Query().Find(&unit, ctx.Request().RouteInt("id")); err != nil || unit.ID == 0 {
		return notFound(ctx, "Unit not found")
	}
	if _, err := facades.Orm().Query().Delete(&unit); err != nil {
		return serverError(ctx, "Failed to delete unit")
	}
	return success(ctx, "Unit deleted successfully", http.Json{})
}
