package feature

import (
	"strings"
	"sync"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
	"goravel/database/seeders"
	"goravel/tests"
)

// HardeningTestSuite exercises the Phase 8 guarantees end-to-end against the
// booted API and a freshly seeded database: authentication, RBAC enforcement,
// security headers, and the atomicity of checkout.
//
// These tests require a reachable MySQL (the same one docker-compose provisions)
// because the suite boots the full application and refreshes the schema. Run
// them with the dev stack up:
//
//	docker compose up -d mysql && (cd backend && go test ./tests/...)
type HardeningTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestHardeningTestSuite(t *testing.T) {
	suite.Run(t, new(HardeningTestSuite))
}

// SetupTest resets the schema and reseeds the demo data before each test so
// every case starts from the known fixture state.
func (s *HardeningTestSuite) SetupTest() {
	// Refresh the schema, then seed explicitly. RefreshDatabase runs
	// `migrate:refresh`, which only seeds when the boolean `--seed` flag is set —
	// passing seeders to it alone does NOT seed. Seed() runs `db:seed` directly.
	s.RefreshDatabase()
	s.Seed(&seeders.DatabaseSeeder{})
}

const jsonContentType = "application/json"

// tokenCache keeps one JWT per demo user for the whole test binary. The demo
// users are reseeded with stable ids on every RefreshDatabase, so a token minted
// once stays valid across cases — and logging in only once per role keeps the
// suite comfortably under the login rate limit.
var (
	tokenCache = map[string]string{}
	tokenMu    sync.Mutex
)

// token returns a bearer token for the given demo user, logging in once and
// caching the result.
func (s *HardeningTestSuite) token(email string) string {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	if t, ok := tokenCache[email]; ok {
		return t
	}

	body := `{"email":"` + email + `","password":"password"}`
	resp, err := s.Http(s.T()).WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/auth/login", strings.NewReader(body))
	s.Require().NoError(err)

	j, err := resp.Json()
	s.Require().NoError(err)
	data, ok := j["data"].(map[string]any)
	s.Require().True(ok, "login response missing data object")
	tok, _ := data["token"].(string)
	s.Require().NotEmpty(tok, "login did not return a token")

	tokenCache[email] = tok
	return tok
}

// --- Authentication ------------------------------------------------------

func (s *HardeningTestSuite) TestLoginSucceedsWithValidCredentials() {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/auth/login", strings.NewReader(`{"email":"admin@kasirku.test","password":"password"}`))
	s.Require().NoError(err)
	resp.AssertOk()

	j, err := resp.Json()
	s.Require().NoError(err)
	data, _ := j["data"].(map[string]any)
	s.NotEmpty(data["token"])
}

func (s *HardeningTestSuite) TestLoginRejectsWrongPassword() {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/auth/login", strings.NewReader(`{"email":"admin@kasirku.test","password":"nope"}`))
	s.Require().NoError(err)
	resp.AssertUnauthorized()
}

func (s *HardeningTestSuite) TestLoginValidatesInput() {
	resp, err := s.Http(s.T()).WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/auth/login", strings.NewReader(`{"password":"password"}`))
	s.Require().NoError(err)
	resp.AssertUnprocessableEntity()
}

// --- Security headers ----------------------------------------------------

func (s *HardeningTestSuite) TestSecurityHeadersOnEveryResponse() {
	resp, err := s.Http(s.T()).Get("/api/v1/health")
	s.Require().NoError(err)
	resp.AssertOk()
	resp.AssertHeader("X-Content-Type-Options", "nosniff")
	resp.AssertHeader("X-Frame-Options", "DENY")
	resp.AssertHeader("Referrer-Policy", "no-referrer")
}

// --- RBAC ----------------------------------------------------------------

func (s *HardeningTestSuite) TestProtectedRouteRequiresAuthentication() {
	resp, err := s.Http(s.T()).Get("/api/v1/products")
	s.Require().NoError(err)
	resp.AssertUnauthorized()
}

func (s *HardeningTestSuite) TestCashierCannotWriteMasterData() {
	// Product writes are limited to admin + warehouse; a cashier must be denied.
	resp, err := s.Http(s.T()).WithToken(s.token("cashier@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/products", strings.NewReader(`{"name":"Should Not Save"}`))
	s.Require().NoError(err)
	resp.AssertForbidden()
}

func (s *HardeningTestSuite) TestWarehouseCannotCheckout() {
	// Checkout is limited to admin + cashier; a warehouse user must be denied.
	resp, err := s.Http(s.T()).WithToken(s.token("warehouse@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/orders", strings.NewReader(`{}`))
	s.Require().NoError(err)
	resp.AssertForbidden()
}

// --- Checkout atomicity --------------------------------------------------

// TestCheckoutSucceedsAndDecrementsStock verifies the happy path drops stock by
// exactly the quantity sold.
func (s *HardeningTestSuite) TestCheckoutSucceedsAndDecrementsStock() {
	// Product 1 (SKU-001) is seeded with 100 units in warehouse 1 at 3500 each.
	body := `{"warehouse_id":1,"items":[{"product_id":1,"quantity":3}],` +
		`"payments":[{"payment_method_id":1,"amount_paid":100000}]}`
	resp, err := s.Http(s.T()).WithToken(s.token("cashier@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/orders", strings.NewReader(body))
	s.Require().NoError(err)
	resp.AssertCreated()

	s.Equal(97, s.stockOf(1, 1))
}

// TestCheckoutRollsBackOnInsufficientStock is the core invariant: if any line
// in the cart cannot be fulfilled, the entire transaction rolls back — earlier
// lines are not decremented and no order is persisted.
func (s *HardeningTestSuite) TestCheckoutRollsBackOnInsufficientStock() {
	// Line 1 (product 1, qty 1) is fine; line 2 (product 2, qty 999999) exceeds
	// its 200-unit stock. Payment is set high enough to clear the payment check
	// so the request reaches the stock movements and fails there.
	body := `{"warehouse_id":1,"items":[` +
		`{"product_id":1,"quantity":1},` +
		`{"product_id":2,"quantity":999999}],` +
		`"payments":[{"payment_method_id":1,"amount_paid":2000000000}]}`
	resp, err := s.Http(s.T()).WithToken(s.token("cashier@kasirku.test")).
		WithHeader("Content-Type", jsonContentType).
		Post("/api/v1/orders", strings.NewReader(body))
	s.Require().NoError(err)
	resp.AssertUnprocessableEntity()

	// Product 1's stock must be untouched despite being the first, valid line.
	s.Equal(100, s.stockOf(1, 1))

	// And no order row may survive the rolled-back transaction.
	count, err := facades.Orm().Query().Model(&models.Order{}).Count()
	s.Require().NoError(err)
	s.Equal(int64(0), count)
}

// stockOf returns the current on-hand quantity for a product in a warehouse.
func (s *HardeningTestSuite) stockOf(productID, warehouseID uint) int {
	var stock models.Stock
	err := facades.Orm().Query().
		Where("product_id", productID).
		Where("warehouse_id", warehouseID).
		First(&stock)
	s.Require().NoError(err)
	return stock.Quantity
}
