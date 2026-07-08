package model

import (
	"testing"
	"time"
)

func TestDraw_Fields(t *testing.T) {
	now := time.Now()
	d := Draw{
		Issue:        "24180",
		DrawTime:     now,
		FrontNumbers: [5]int{5, 12, 18, 23, 31},
		BackNumbers:  [2]int{7, 11},
		SaleAmount:   310000000,
		PoolAmount:   920000000,
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"期号", d.Issue, "24180"},
		{"前区号码数量", len(d.FrontNumbers), 5},
		{"后区号码数量", len(d.BackNumbers), 2},
		{"第一个前区号码", d.FrontNumbers[0], 5},
		{"第一个后区号码", d.BackNumbers[0], 7},
		{"销售额", d.SaleAmount, int64(310000000)},
		{"奖池金额", d.PoolAmount, int64(920000000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestDraw_DefaultZero(t *testing.T) {
	var d Draw
	if d.Issue != "" {
		t.Errorf("expected empty issue, got %s", d.Issue)
	}
	if !d.DrawTime.IsZero() {
		t.Errorf("expected zero time")
	}
}
