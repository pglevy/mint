package jsnum

import (
	"math"
	"testing"
)

func TestFormat(t *testing.T) {
	// Expected values are what JavaScript's String(n) prints.
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{math.Copysign(0, -1), "0"},
		{1, "1"},
		{-1, "-1"},
		{16, "16"},
		{100, "100"},
		{3.14, "3.14"},
		{0.1, "0.1"},
		{0.30000000000000004, "0.30000000000000004"}, // 0.1 + 0.2 in float64
		{-2.5, "-2.5"},
		{0.000001, "0.000001"},
		{1e-7, "1e-7"},
		{1.5e-7, "1.5e-7"},
		{123456789012345680000, "123456789012345680000"},
		{1e21, "1e+21"},
		{1.5e21, "1.5e+21"},
		{1e300, "1e+300"},
		{5e-324, "5e-324"},
		{math.MaxFloat64, "1.7976931348623157e+308"},
		{9007199254740993, "9007199254740992"},
		{math.Inf(1), "Infinity"},
		{math.Inf(-1), "-Infinity"},
		{math.NaN(), "NaN"},
	}
	for _, tt := range tests {
		if got := Format(tt.in); got != tt.want {
			t.Errorf("Format(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		in   string
		want float64 // NaN means "expect NaN"
	}{
		{"16px", 16},
		{"0.5rem", 0.5},
		{"  12px", 12},
		{"\u00a07", 7},
		{"\ufeff8", 8},
		{"-3", -3},
		{"+4", 4},
		{".5em", 0.5},
		{"5.", 5},
		{"1e3x", 1000},
		{"1e", 1},
		{"3.14abc", 3.14},
		{"Infinity", math.Inf(1)},
		{"-Infinityx", math.Inf(-1)},
		{"1e400", math.Inf(1)},
		{"#fff", math.NaN()},
		{"", math.NaN()},
		{"auto", math.NaN()},
		{"e5", math.NaN()},
		{"0x10", 0},
		{"1_000", 1},
	}
	for _, tt := range tests {
		got := ParseFloat(tt.in)
		if math.IsNaN(tt.want) != math.IsNaN(got) || (!math.IsNaN(got) && got != tt.want) {
			t.Errorf("ParseFloat(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestToNumber(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"", 0},
		{"  ", 0},
		{"0.5", 0.5},
		{" 1 ", 1},
		{"-.5", -0.5},
		{"5.", 5},
		{"1e3", 1000},
		{"0x1F", 31},
		{"0o17", 15},
		{"0b101", 5},
		{"Infinity", math.Inf(1)},
		{"-Infinity", math.Inf(-1)},
		{"16px", math.NaN()},
		{"1e", math.NaN()},
		{".", math.NaN()},
		{"0x", math.NaN()},
		{"0x-1", math.NaN()},
		{"1_000", math.NaN()},
		{"infinity", math.NaN()},
		{"NaN", math.NaN()},
	}
	for _, tt := range tests {
		got := ToNumber(tt.in)
		if math.IsNaN(tt.want) != math.IsNaN(got) || (!math.IsNaN(got) && got != tt.want) {
			t.Errorf("ToNumber(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestRound(t *testing.T) {
	// Math.round rounds halves toward +Infinity.
	tests := []struct{ in, want float64 }{
		{0.4, 0}, {0.5, 1}, {1.5, 2}, {2.5, 3}, {-0.5, 0}, {-1.5, -1}, {-2.5, -2},
		{0.49999999999999994, 0}, {57.49999999999999, 57}, {1e300, 1e300},
	}
	for _, tt := range tests {
		if got := Round(tt.in); got != tt.want {
			t.Errorf("Round(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
	if got := Round(math.NaN()); !math.IsNaN(got) {
		t.Errorf("Round(NaN) = %v, want NaN", got)
	}
}
