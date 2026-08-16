package feature

import (
	"strconv"
	"strings"

	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

// This file adds inventory/POS flow coverage to HardeningTestSuite (defined in
// hardening_test.go): the stock engine's ledger, order void returning stock,
// barcode lookup, and payment validation. It reuses that suite's per-role token
// cache and SetupTest (RefreshDatabase), so every case starts from seeded data.

// createOrder posts a checkout as the cashier and returns the new order's id.
func (s *HardeningTestSuite) createOrder(body string) uint {
	resp, err := s.Http(s.T()).WithToken(s.token("cashier@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/orders", strings.NewReader(body))
	s.Require().NoError(err)
	resp.AssertCreated()

	j, err := resp.Json()
	s.Require().NoError(err)
	data, ok := j["data"].(map[string]any)
	s.Require().True(ok, "checkout response missing data object")
	id, ok := data["id"].(float64) // JSON numbers decode as float64
	s.Require().True(ok, "checkout response missing order id")
	return uint(id)
}

// --- Void returns stock --------------------------------------------------

// TestVoidReturnsStockToInventory checks that voiding an order writes a RETURN
// movement that restores exactly what the sale removed.
func (s *HardeningTestSuite) TestVoidReturnsStockToInventory() {
	// Product 1 starts at 100. Sell 5 → 95, then void → back to 100.
	orderID := s.createOrder(`{"warehouse_id":1,"items":[{"product_id":1,"quantity":5}],` +
		`"payments":[{"payment_method_id":1,"amount_paid":100000}]}`)
	s.Require().Equal(95, s.stockOf(1, 1))

	// Void is admin-only.
	resp, err := s.Http(s.T()).WithToken(s.token("admin@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/orders/"+strconv.FormatUint(uint64(orderID), 10)+"/void", strings.NewReader(`{}`))
	s.Require().NoError(err)
	resp.AssertOk()

	s.Equal(100, s.stockOf(1, 1))

	// The order is now cancelled.
	var order models.Order
	err = facades.Orm().Query().Where("id", orderID).First(&order)
	s.Require().NoError(err)
	s.Equal("cancelled", order.Status)
}

// TestCashierCannotVoid confirms void stays admin-only.
func (s *HardeningTestSuite) TestCashierCannotVoid() {
	orderID := s.createOrder(`{"warehouse_id":1,"items":[{"product_id":1,"quantity":1}],` +
		`"payments":[{"payment_method_id":1,"amount_paid":100000}]}`)

	resp, err := s.Http(s.T()).WithToken(s.token("cashier@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/orders/"+strconv.FormatUint(uint64(orderID), 10)+"/void", strings.NewReader(`{}`))
	s.Require().NoError(err)
	resp.AssertForbidden()
}

// --- Stock-in ledger -----------------------------------------------------

// TestStockInIncrementsStockAndWritesMovement exercises the receiving flow: a
// warehouse user records a stock-in, stock goes up by the received quantity, and
// a STOCK_IN ledger row is written with matching before/after quantities.
func (s *HardeningTestSuite) TestStockInIncrementsStockAndWritesMovement() {
	before := s.stockOf(1, 1) // seeded 100

	body := `{"warehouse_id":1,"items":[{"product_id":1,"quantity":30,"purchase_price":2000}]}`
	resp, err := s.Http(s.T()).WithToken(s.token("warehouse@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/stock-ins", strings.NewReader(body))
	s.Require().NoError(err)
	resp.AssertCreated()

	s.Equal(before+30, s.stockOf(1, 1))

	var movement models.StockMovement
	err = facades.Orm().Query().
		Where("product_id", 1).
		Where("type", models.MovementStockIn).
		OrderByDesc("id").First(&movement)
	s.Require().NoError(err)
	s.Equal(30, movement.Quantity)
	s.Equal(before, movement.QuantityBefore)
	s.Equal(before+30, movement.QuantityAfter)
}

// --- POS barcode lookup --------------------------------------------------

func (s *HardeningTestSuite) TestBarcodeScanReturnsProduct() {
	// 8991234500011 is seeded as SKU-001 (Air Mineral 600ml).
	resp, err := s.Http(s.T()).WithToken(s.token("cashier@kasirku.test")).
		Get("/api/v1/pos/scan/8991234500011")
	s.Require().NoError(err)
	resp.AssertOk()

	j, err := resp.Json()
	s.Require().NoError(err)
	data, _ := j["data"].(map[string]any)
	s.Equal("SKU-001", data["sku"])
}

func (s *HardeningTestSuite) TestBarcodeScanMissingReturns404() {
	resp, err := s.Http(s.T()).WithToken(s.token("cashier@kasirku.test")).
		Get("/api/v1/pos/scan/does-not-exist")
	s.Require().NoError(err)
	resp.AssertNotFound()
}

// --- Payment validation --------------------------------------------------

// TestCheckoutRejectsInsufficientPayment ensures an under-paid cart is refused
// and leaves stock untouched.
func (s *HardeningTestSuite) TestCheckoutRejectsInsufficientPayment() {
	// 3 × 3500 = 10500 due, but only 5000 tendered.
	body := `{"warehouse_id":1,"items":[{"product_id":1,"quantity":3}],` +
		`"payments":[{"payment_method_id":1,"amount_paid":5000}]}`
	resp, err := s.Http(s.T()).WithToken(s.token("cashier@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/orders", strings.NewReader(body))
	s.Require().NoError(err)
	resp.AssertUnprocessableEntity()

	s.Equal(100, s.stockOf(1, 1))
}
