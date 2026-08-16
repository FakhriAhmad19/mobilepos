package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// CategoryController — CRUD for product categories (PRD §8.5).
type CategoryController struct{}

func NewCategoryController() *CategoryController {
	return &CategoryController{}
}

func (r *CategoryController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.Category{})
	if s := ctx.Request().Query("search"); s != "" {
		query = query.Where("name LIKE ?", "%"+s+"%")
	}
	if status := ctx.Request().Query("status"); status != "" {
		query = query.Where("status", status)
	}

	var categories []models.Category
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &categories, &total); err != nil {
		return serverError(ctx, "Failed to fetch categories")
	}

	return successMeta(ctx, "Categories retrieved successfully", categories, paginationMeta(page, perPage, total))
}

func (r *CategoryController) Show(ctx http.Context) http.Response {
	var category models.Category
	if err := facades.Orm().Query().Find(&category, ctx.Request().RouteInt("id")); err != nil || category.ID == 0 {
		return notFound(ctx, "Category not found")
	}
	return success(ctx, "Category retrieved successfully", category)
}

func (r *CategoryController) Store(ctx http.Context) http.Response {
	validator, err := ctx.Request().Validate(map[string]any{
		"name": "required|max_len:100",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	category := models.Category{
		Name:        ctx.Request().Input("name"),
		Description: ctx.Request().Input("description"),
		Status:      ctx.Request().Input("status", "active"),
	}
	if err := facades.Orm().Query().Create(&category); err != nil {
		return serverError(ctx, "Failed to create category")
	}
	return created(ctx, "Category created successfully", category)
}

func (r *CategoryController) Update(ctx http.Context) http.Response {
	var category models.Category
	if err := facades.Orm().Query().Find(&category, ctx.Request().RouteInt("id")); err != nil || category.ID == 0 {
		return notFound(ctx, "Category not found")
	}

	validator, err := ctx.Request().Validate(map[string]any{
		"name": "required|max_len:100",
	})
	if err != nil {
		return serverError(ctx, "Validation error")
	}
	if validator.Fails() {
		return validationFailed(ctx, validator.Errors().All())
	}

	category.Name = ctx.Request().Input("name", category.Name)
	category.Description = ctx.Request().Input("description", category.Description)
	category.Status = ctx.Request().Input("status", category.Status)
	if err := facades.Orm().Query().Save(&category); err != nil {
		return serverError(ctx, "Failed to update category")
	}
	return success(ctx, "Category updated successfully", category)
}

func (r *CategoryController) Destroy(ctx http.Context) http.Response {
	var category models.Category
	if err := facades.Orm().Query().Find(&category, ctx.Request().RouteInt("id")); err != nil || category.ID == 0 {
		return notFound(ctx, "Category not found")
	}
	if _, err := facades.Orm().Query().Delete(&category); err != nil {
		return serverError(ctx, "Failed to delete category")
	}
	return success(ctx, "Category deleted successfully", http.Json{})
}
