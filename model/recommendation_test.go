package model

import (
	"testing"
)

func TestRecommendNumber_Fields(t *testing.T) {
	rn := RecommendNumber{
		Number: 5,
		Reason: "🔥热",
	}

	if rn.Number != 5 {
		t.Errorf("expected number 5, got %d", rn.Number)
	}
	if rn.Reason != "🔥热" {
		t.Errorf("expected reason 🔥热, got %s", rn.Reason)
	}
}

func TestRecommendation_Fields(t *testing.T) {
	rec := Recommendation{
		FrontNumbers: []RecommendNumber{
			{Number: 1, Reason: "🔥热"},
			{Number: 2, Reason: "🔥热"},
			{Number: 3, Reason: "🌡️温"},
			{Number: 4, Reason: "🌡️温"},
			{Number: 5, Reason: "📊遗漏"},
		},
		BackNumbers: []RecommendNumber{
			{Number: 6, Reason: "🔥热"},
			{Number: 7, Reason: "📊遗漏"},
		},
	}

	if len(rec.FrontNumbers) != 5 {
		t.Errorf("expected 5 front numbers, got %d", len(rec.FrontNumbers))
	}
	if len(rec.BackNumbers) != 2 {
		t.Errorf("expected 2 back numbers, got %d", len(rec.BackNumbers))
	}
}
