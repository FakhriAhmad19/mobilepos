package services

import "testing"

func TestLineSubtotal(t *testing.T) {
	cases := []struct {
		name     string
		quantity int
		price    float64
		discount float64
		want     float64
	}{
		{"no discount", 3, 1000, 0, 3000},
		{"with discount", 2, 1500, 500, 2500},
		{"discount larger than line is floored at zero", 1, 1000, 5000, 0},
		{"zero quantity", 0, 1000, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := LineSubtotal(c.quantity, c.price, c.discount); got != c.want {
				t.Errorf("LineSubtotal(%d, %v, %v) = %v, want %v",
					c.quantity, c.price, c.discount, got, c.want)
			}
		})
	}
}

func TestGrandTotal(t *testing.T) {
	cases := []struct {
		name     string
		subtotal float64
		discount float64
		tax      float64
		want     float64
	}{
		{"discount and tax", 10000, 1000, 1100, 10100},
		{"no adjustments", 5000, 0, 0, 5000},
		{"over-discount is floored at zero", 1000, 5000, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := GrandTotal(c.subtotal, c.discount, c.tax); got != c.want {
				t.Errorf("GrandTotal(%v, %v, %v) = %v, want %v",
					c.subtotal, c.discount, c.tax, got, c.want)
			}
		})
	}
}

func TestPaymentCovers(t *testing.T) {
	cases := []struct {
		name       string
		totalPaid  float64
		grandTotal float64
		want       bool
	}{
		{"exact payment", 10000, 10000, true},
		{"overpayment", 20000, 10000, true},
		{"underpayment", 9000, 10000, false},
		{"rounding tolerance covers a hair short", 9999.9999, 10000, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := PaymentCovers(c.totalPaid, c.grandTotal); got != c.want {
				t.Errorf("PaymentCovers(%v, %v) = %v, want %v",
					c.totalPaid, c.grandTotal, got, c.want)
			}
		})
	}
}
