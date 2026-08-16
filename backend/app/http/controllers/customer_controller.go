package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// CustomerController — CRUD for customers (PRD §8.7).
type CustomerController struct{}

func NewCustomerController() *CustomerController {
	return &CustomerController{}
}

func (r *CustomerController) Index(ctx http.Context) http.Response {
	page, perPage := paginationParams(ctx)

	query := facades.Orm().Query().Model(&models.Customer{})
	if s := ctx.Request().Query("search"); s != "" {
		like := "%" + s + "%"
		query = query.Where("name LIKE ? OR phone LIKE ? OR email LIKE ?", like, like, like)
	}
	if status := ctx.Request().Query("status"); status != "" {
		query = query.Where("status", status)
	}

	var customers []models.Customer
	var total int64
	if err := query.OrderByDesc("id").Paginate(page, perPage, &customers, &total); err != nil {
		return serverError(ctx, "Failed to fetch customers")
	}
	return successMeta(ctx, "Customers retrieved successfully", customers, paginationMeta(page, perPage, total))
}

func (r *CustomerController) Show(ctx http.Context) http.Response {
	var customer models.Customer
	if err := facades.Orm().Query().Find(&customer, ctx.Request().RouteInt("id")); err != nil || customer.ID == 0 {
		return notFound(ctx, "Customer not found")
	}
	return success(ctx, "Customer retrieved successfully", customer)
}

func (r *CustomerController) Store(ctx http.Context) http.Response {
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

	customer := models.Customer{
		Name:    ctx.Request().Input("name"),
		Phone:   ctx.Request().Input("phone"),
		Email:   ctx.Request().Input("email"),
		Address: ctx.Request().Input("address"),
		Status:  ctx.Request().Input("status", "active"),
	}
	if err := facades.Orm().Query().Create(&customer); err != nil {
		return serverError(ctx, "Failed to create customer")
	}
	return created(ctx, "Customer created successfully", customer)
}

func (r *CustomerController) Update(ctx http.Context) http.Response {
	var customer models.Customer
	if err := facades.Orm().Query().Find(&customer, ctx.Request().RouteInt("id")); err != nil || customer.ID == 0 {
		return notFound(ctx, "Customer not found")
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

	customer.Name = ctx.Request().Input("name", customer.Name)
	customer.Phone = ctx.Request().Input("phone", customer.Phone)
	customer.Email = ctx.Request().Input("email", customer.Email)
	customer.Address = ctx.Request().Input("address", customer.Address)
	customer.Status = ctx.Request().Input("status", customer.Status)
	if err := facades.Orm().Query().Save(&customer); err != nil {
		return serverError(ctx, "Failed to update customer")
	}
	return success(ctx, "Customer updated successfully", customer)
}

func (r *CustomerController) Destroy(ctx http.Context) http.Response {
	var customer models.Customer
	if err := facades.Orm().Query().Find(&customer, ctx.Request().RouteInt("id")); err != nil || customer.ID == 0 {
		return notFound(ctx, "Customer not found")
	}
	if _, err := facades.Orm().Query().Delete(&customer); err != nil {
		return serverError(ctx, "Failed to delete customer")
	}
	return success(ctx, "Customer deleted successfully", http.Json{})
}
