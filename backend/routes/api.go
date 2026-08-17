package routes

import (
	nethttp "net/http"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"

	"goravel/app/facades"
	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

// crudController is the method set shared by all master-data controllers.
type crudController interface {
	Index(http.Context) http.Response
	Show(http.Context) http.Response
	Store(http.Context) http.Response
	Update(http.Context) http.Response
	Destroy(http.Context) http.Response
}

// resource registers RESTful routes: reads are open to any authenticated user,
// writes are restricted to the given roles (PRD §8.2 RBAC).
func resource(router route.Router, name string, c crudController, writeRoles ...string) {
	router.Get(name, c.Index)
	router.Get(name+"/{id}", c.Show)
	router.Middleware(middleware.Role(writeRoles...)).Post(name, c.Store)
	router.Middleware(middleware.Role(writeRoles...)).Put(name+"/{id}", c.Update)
	router.Middleware(middleware.Role(writeRoles...)).Delete(name+"/{id}", c.Destroy)
}

// Api registers the versioned REST API surface described in the PRD (§11).
func Api() {
	healthController := controllers.NewHealthController()
	authController := controllers.NewAuthController()

	// Baseline security headers + CORS for the API on every response
	// (Phase 8 — security; CORS lets browser clients call the API cross-origin).
	facades.Route().GlobalMiddleware(middleware.SecureHeaders(), middleware.Cors())

	// Turn any unhandled panic into the PRD error envelope instead of leaking a
	// stack trace or dropping the connection (Phase 8 — consistent errors).
	facades.Route().Recover(func(ctx http.Context, err any) {
		facades.Log().Error(err)
		_ = ctx.Response().Json(nethttp.StatusInternalServerError, http.Json{
			"success": false,
			"message": "Internal server error",
		}).Abort()
	})

	// Brute-force protection on login: 10 attempts / minute / client IP.
	loginThrottle := middleware.Throttle("login", 10, time.Minute)

	facades.Route().Prefix("api/v1").Group(func(router route.Router) {
		// --- Public ---------------------------------------------------
		router.Get("health", healthController.Show)
		router.Get("ping", func(ctx http.Context) http.Response {
			return ctx.Response().Success().Json(http.Json{
				"success": true, "message": "pong", "data": http.Json{}, "meta": http.Json{},
			})
		})
		router.Middleware(loginThrottle).Post("auth/login", authController.Login)

		// --- Authenticated (any role) --------------------------------
		router.Middleware(middleware.Authenticate()).Group(func(auth route.Router) {
			auth.Get("auth/me", authController.Me)
			auth.Post("auth/logout", authController.Logout)
			auth.Post("auth/refresh", authController.Refresh)
			auth.Post("auth/change-password", authController.ChangePassword)

			// --- Master data (PRD §8.4–§8.8, Phase 4) ----------------
			// Reads: any authenticated user. Writes: roles noted per resource.
			resource(auth, "categories", controllers.NewCategoryController(), "admin", "warehouse")
			resource(auth, "units", controllers.NewUnitController(), "admin", "warehouse")
			resource(auth, "products", controllers.NewProductController(), "admin", "warehouse")
			resource(auth, "suppliers", controllers.NewSupplierController(), "admin", "warehouse")
			resource(auth, "warehouses", controllers.NewWarehouseController(), "admin", "warehouse")
			resource(auth, "customers", controllers.NewCustomerController(), "admin", "cashier")

			// --- Inventory (PRD §8.9–§8.14, Phase 5) -----------------
			// Reads: any authenticated user. Writes: admin + warehouse.
			auth.Get("inventory", controllers.NewInventoryController().Index)
			auth.Get("stock-movements", controllers.NewStockMovementController().Index)

			stockIn := controllers.NewStockInController()
			auth.Get("stock-ins", stockIn.Index)
			auth.Get("stock-ins/{id}", stockIn.Show)

			stockOut := controllers.NewStockOutController()
			auth.Get("stock-outs", stockOut.Index)
			auth.Get("stock-outs/{id}", stockOut.Show)

			adjustment := controllers.NewStockAdjustmentController()
			auth.Get("stock-adjustments", adjustment.Index)
			auth.Get("stock-adjustments/{id}", adjustment.Show)

			opname := controllers.NewStockOpnameController()
			auth.Get("stock-opnames", opname.Index)
			auth.Get("stock-opnames/{id}", opname.Show)

			auth.Middleware(middleware.Role("admin", "warehouse")).Group(func(wh route.Router) {
				wh.Post("stock-ins", stockIn.Store)
				wh.Post("stock-outs", stockOut.Store)
				wh.Post("stock-adjustments", adjustment.Store)
				wh.Post("stock-opnames", opname.Store)
				wh.Post("stock-opnames/{id}/complete", opname.Complete)
			})

			// --- POS / transactions (PRD §8.15–§8.22, Phase 6) -------
			auth.Get("payment-methods", controllers.NewPaymentMethodController().Index)
			// Kept off the `products/{id}` path to avoid static/param route conflicts.
			auth.Get("pos/scan/{barcode}", controllers.NewProductController().Barcode)

			orderController := controllers.NewOrderController()
			auth.Get("orders", orderController.Index)
			auth.Get("orders/{id}", orderController.Show)
			// Checkout: cashiers and admins.
			auth.Middleware(middleware.Role("admin", "cashier")).Post("orders", orderController.Store)
			// Void/refund: admin only.
			auth.Middleware(middleware.Role("admin")).Post("orders/{id}/void", orderController.Void)

			// --- Dashboard & reports (PRD §8.23–§8.24, Phase 7) ------
			auth.Get("dashboard", controllers.NewDashboardController().Show)

			report := controllers.NewReportController()
			auth.Middleware(middleware.Role("admin")).Group(func(rep route.Router) {
				rep.Get("reports/sales", report.Sales)
				rep.Get("reports/products", report.Products)
				rep.Get("reports/inventory", report.Inventory)
				rep.Get("reports/cashiers", report.Cashiers)
			})
		})
	})
}
