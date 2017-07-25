package rep

import (
	"math"
	"testing"
)

func TestIsMainBuilding(t *testing.T) {
	cases := []struct {
		name   string
		isMain bool
	}{
		{"Nexus", true},
		{"CommandCenter", true},
		{"Hatchery", true},
		{"", false},
		{"nexus", false},
		{"kitty", false},
	}

	for _, c := range cases {
		if got := isMainBuilding(c.name); got != c.isMain {
			t.Errorf("Expected: %v, got: %v", c.isMain, got)
		}
	}
}

func TestAngleToClock(t *testing.T) {
	cases := []struct {
		angle float64
		clock int
	}{
		{0, 3},
		{math.Pi / 2, 12},
		{math.Pi, 9},
		{math.Pi * 3 / 2, 6},

		{0 + math.Pi*6, 3},
		{0 - math.Pi*6, 3},

		{math.Pi/2 + math.Pi/13, 12},
		{math.Pi/2 - math.Pi/13, 12},
	}

	for _, c := range cases {
		if got := angleToClock(c.angle); got != c.clock {
			t.Errorf("Expected: %v, got: %v", c.clock, got)
		}
	}
}

func TestCalcSQ(t *testing.T) {
	cases := []struct {
		unspent, income int64
		sq              int
	}{
		// Test table taken from http://www.teamliquid.net/forum/starcraft-2/266019-do-you-macro-like-a-pro
		{959, 1970, 94},
		{408, 1358, 95},
		{263, 844, 85},
		{447, 1459, 96},
		{585, 1849, 106},
		{262, 901, 88},
		{814, 1734, 89},
		{442, 1486, 98},
		{696, 1682, 92},
		{1641, 2107, 82},
		{670, 1291, 74},
		{2201, 2148, 74},
		{584, 1540, 91},
		{2222, 2085, 70},
		{556, 1486, 90},
		{1580, 2078, 82},
		{547, 1219, 78},
	}

	for _, c := range cases {
		if got := calcSQ(c.unspent, c.income); got != c.sq {
			t.Errorf("Expected: %v, got: %v", c.sq, got)
		}
	}
}
