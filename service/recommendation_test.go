package service

import (
	"context"
	"testing"

	"github.com/user/lottery/model"
)

func TestRecommendService_GenerateRecommendation(t *testing.T) {
	stats := &model.Statistics{
		FrontHot:  []model.NumberFrequency{{Number: 1, Count: 15}, {Number: 2, Count: 14}, {Number: 3, Count: 13}},
		FrontWarm: []model.NumberFrequency{{Number: 10, Count: 8}, {Number: 11, Count: 7}},
		FrontCold: []model.NumberFrequency{{Number: 30, Count: 2, MissValue: 20}, {Number: 31, Count: 1, MissValue: 25}},
		BackHot:   []model.NumberFrequency{{Number: 6, Count: 10}, {Number: 7, Count: 9}},
		BackWarm:  []model.NumberFrequency{{Number: 8, Count: 5}},
		BackCold:  []model.NumberFrequency{{Number: 12, Count: 1, MissValue: 15}},
	}

	svc := NewRecommendService(nil)
	rec, err := svc.GenerateRecommendation(context.Background(), stats)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rec.FrontNumbers) != 5 {
		t.Errorf("expected 5 front numbers, got %d", len(rec.FrontNumbers))
	}
	if len(rec.BackNumbers) != 2 {
		t.Errorf("expected 2 back numbers, got %d", len(rec.BackNumbers))
	}

	seen := make(map[int]bool)
	for _, n := range rec.FrontNumbers {
		if seen[n.Number] {
			t.Errorf("duplicate front number: %d", n.Number)
		}
		seen[n.Number] = true
	}
	for _, n := range rec.BackNumbers {
		if seen[n.Number] {
			t.Errorf("duplicate back number: %d", n.Number)
		}
		seen[n.Number] = true
	}
}

func TestRecommendService_GenerateRecommendation_NilStats(t *testing.T) {
	svc := NewRecommendService(nil)
	_, err := svc.GenerateRecommendation(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil stats")
	}
}

func TestPickWeighted_EmptyInput(t *testing.T) {
	result := pickWeighted([]model.NumberFrequency{}, 3, map[int]bool{})
	if result != nil {
		t.Errorf("expected nil for empty input")
	}
}

func TestPickMissed_EmptyInput(t *testing.T) {
	result := pickMissed([]model.NumberFrequency{}, 3, map[int]bool{})
	if result != nil {
		t.Errorf("expected nil for empty input")
	}
}

func TestPickWeighted_ExcludedAll(t *testing.T) {
	freqs := []model.NumberFrequency{{Number: 1, Count: 10}}
	excluded := map[int]bool{1: true}
	result := pickWeighted(freqs, 1, excluded)
	if result != nil {
		t.Errorf("expected nil when all excluded")
	}
}

func TestPickMissed_ExcludedAll(t *testing.T) {
	freqs := []model.NumberFrequency{{Number: 1, Count: 10}}
	excluded := map[int]bool{1: true}
	result := pickMissed(freqs, 1, excluded)
	if len(result) != 0 {
		t.Errorf("expected 0 when all excluded")
	}
}
