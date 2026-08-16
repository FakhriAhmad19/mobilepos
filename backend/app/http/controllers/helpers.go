package controllers

import (
	nethttp "net/http"

	"github.com/goravel/framework/contracts/http"
)

// success writes a 200 response using the PRD envelope (§11).
func success(ctx http.Context, message string, data any) http.Response {
	return ctx.Response().Success().Json(http.Json{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    http.Json{},
	})
}

// successMeta writes a 200 response including pagination meta.
func successMeta(ctx http.Context, message string, data any, meta http.Json) http.Response {
	return ctx.Response().Success().Json(http.Json{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

// created writes a 201 response.
func created(ctx http.Context, message string, data any) http.Response {
	return ctx.Response().Json(nethttp.StatusCreated, http.Json{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    http.Json{},
	})
}

// notFound writes a 404 error response.
func notFound(ctx http.Context, message string) http.Response {
	return ctx.Response().Json(nethttp.StatusNotFound, http.Json{
		"success": false,
		"message": message,
	})
}

// unprocessable writes a 422 error response with a single message.
func unprocessable(ctx http.Context, message string) http.Response {
	return ctx.Response().Json(nethttp.StatusUnprocessableEntity, http.Json{
		"success": false,
		"message": message,
	})
}

// paginationParams reads and clamps the page / per_page query parameters.
func paginationParams(ctx http.Context) (page, perPage int) {
	page = ctx.Request().QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	perPage = ctx.Request().QueryInt("per_page", 15)
	if perPage < 1 {
		perPage = 15
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

// paginationMeta builds the pagination block for the response meta.
func paginationMeta(page, perPage int, total int64) http.Json {
	lastPage := 1
	if perPage > 0 {
		lastPage = int((total + int64(perPage) - 1) / int64(perPage))
	}
	if lastPage < 1 {
		lastPage = 1
	}
	return http.Json{
		"current_page": page,
		"per_page":     perPage,
		"total":        total,
		"last_page":    lastPage,
	}
}
