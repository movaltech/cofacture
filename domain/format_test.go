// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package domain

import (
	"testing"

	"github.com/movaltech/decimal"
)

// TestFormatCents locks in the exact 2-decimal, no-thousands-separator format DIAN requires —
// this function is shared by every XML serializer and every CUFE/CUDE/CUDS hash composer in the
// library, so a regression here would silently corrupt both at once.
func TestFormatCents(t *testing.T) {
	cases := []struct {
		cents int64
		want  string
	}{
		{0, "0.00"},
		{5, "0.05"},
		{100, "1.00"},
		{350_000_000, "3500000.00"},
		{-500, "-5.00"},
		{-5, "-0.05"},
	}

	for _, c := range cases {
		if got := FormatCents(c.cents); got != c.want {
			t.Errorf("FormatCents(%d) = %q, want %q", c.cents, got, c.want)
		}
	}
}

// TestFormatFixed locks in exact fixed-decimal rendering for quantities (6 places) and
// percentages (2 places), including the half-way cases where float64 with %.2f rounds the
// wrong way (2.675 is stored as 2.67499999... in binary and printed as "2.67").
func TestFormatFixed(t *testing.T) {
	tests := []struct {
		in     string
		places int32
		want   string
	}{
		{"19", 2, "19.00"},
		{"3.5", 2, "3.50"},
		{"2.675", 2, "2.68"},
		{"1.005", 2, "1.01"},
		{"0.414", 2, "0.41"},
		{"1", 6, "1.000000"},
		{"1234567890.1234565", 6, "1234567890.123457"},
		{"0.0000005", 6, "0.000001"},
		{"-2.675", 2, "-2.68"},
	}
	for _, tt := range tests {
		if got := FormatFixed(decimal.MustParse(tt.in), tt.places); got != tt.want {
			t.Errorf("FormatFixed(%s, %d) = %q, want %q", tt.in, tt.places, got, tt.want)
		}
	}
}

func TestFormatFixedPanicsOnInvalidPlaces(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("FormatFixed with negative places did not panic")
		}
	}()
	FormatFixed(decimal.MustParse("1"), -1)
}
