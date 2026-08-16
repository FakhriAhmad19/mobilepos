package services

// paymentEpsilon is the tolerance used when comparing a client-supplied payment
// total against the server-computed grand total, so floating-point rounding can
// never reject an otherwise-exact payment. Prices themselves stay
// server-authoritative (PRD §18).
const paymentEpsilon = 0.001

// LineSubtotal computes a single order line's subtotal: quantity × unit price
// less the line discount, floored at zero so a discount can never make a line
// negative.
func LineSubtotal(quantity int, unitPrice, discount float64) float64 {
	sub := float64(quantity)*unitPrice - discount
	if sub < 0 {
		return 0
	}
	return sub
}

// GrandTotal applies the order-level discount and tax to the summed line
// subtotal, floored at zero.
func GrandTotal(subtotal, discount, tax float64) float64 {
	total := subtotal - discount + tax
	if total < 0 {
		return 0
	}
	return total
}

// PaymentCovers reports whether the amount paid covers the grand total, within
// paymentEpsilon.
func PaymentCovers(totalPaid, grandTotal float64) bool {
	return totalPaid+paymentEpsilon >= grandTotal
}
