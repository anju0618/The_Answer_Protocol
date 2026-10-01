package main

import "testing"

func TestBarRatioClamps(t *testing.T) {
	cases := []struct {
		cur, max int
		want     float32
	}{
		{50, 100, 0.5}, {100, 100, 1}, {150, 100, 1}, {0, 100, 0}, {-5, 100, 0}, {10, 0, 0},
	}
	for _, c := range cases {
		if got := barRatio(c.cur, c.max); got != c.want {
			t.Errorf("barRatio(%d,%d) = %v, want %v", c.cur, c.max, got, c.want)
		}
	}
}

func TestBarColorTurnsRedWhenLow(t *testing.T) {
	if barColor(0.9) == barColor(0.1) {
		t.Fatal("full and nearly-empty bars must differ in colour")
	}
	if barColor(0.4) == barColor(0.9) || barColor(0.4) == barColor(0.1) {
		t.Fatal("mid range must have its own colour")
	}
}
