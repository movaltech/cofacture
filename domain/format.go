// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package domain

import (
	"fmt"

	"github.com/movaltech/decimal"
)

// FormatCents renders cents as the 2-decimal string DIAN requires (no thousands separator,
// dot as the decimal separator) — used both when serializing XML and when composing the
// CUFE/CUDE hash input strings, which share the same format.
func FormatCents(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	s := fmt.Sprintf("%d.%02d", cents/100, cents%100)
	if neg {
		s = "-" + s
	}
	return s
}

// FormatFixed renders d with exactly places decimals (no thousands separator, dot as the
// decimal separator), rounding half away from zero when d has more decimals than that — e.g.
// FormatFixed(decimal.MustParse("19"), 2) == "19.00" and
// FormatFixed(decimal.MustParse("2.675"), 2) == "2.68". It is exact: unlike formatting a
// float64 with %.2f, the value never goes through binary floating point.
//
// places must be a small non-negative number (DIAN uses 2 and 6); FormatFixed panics
// otherwise, since that can only be a programming error.
func FormatFixed(d decimal.Decimal, places int32) string {
	r, err := d.Rescale(places, decimal.RoundHalfUp)
	if err != nil {
		panic(fmt.Sprintf("domain: FormatFixed(%s, %d): %v", d, places, err))
	}
	return r.String()
}
