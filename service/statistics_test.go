package service

import (
	"context"
	"testing"

	"github.com/user/lottery/client"
	"github.com/user/lottery/client/mock"
	"github.com/user/lottery/model"
)

func TestStatsService_CalculateStats_Success(t *testing.T) {
	mockAPI := &mock.MockLotteryAPI{
		FetchDrawsFunc: func(ctx context.Context, opts ...client.Option) ([]model.Draw, error) {
			return generateTestDraws(50), nil
		},
	}
	svc := NewStatsService(mockAPI)
	draws := generateTestDraws(50)
	stats, err := svc.CalculateStats(context.Background(), draws, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats.FrontHot) == 0 {
		t.Errorf("expected non-empty front hot numbers")
	}
	if len(stats.BackHot) == 0 {
		t.Errorf("expected non-empty back hot numbers")
	}
}

func TestStatsService_CalculateStats_EmptyDraws(t *testing.T) {
	mockAPI := &mock.MockLotteryAPI{}
	svc := NewStatsService(mockAPI)
	_, err := svc.CalculateStats(context.Background(), []model.Draw{}, 20)
	if err == nil {
		t.Fatal("expected error for empty draws")
	}
}

func TestStatsService_CalculateStats_InvalidRange(t *testing.T) {
	mockAPI := &mock.MockLotteryAPI{}
	svc := NewStatsService(mockAPI)
	_, err := svc.CalculateStats(context.Background(), generateTestDraws(10), 0)
	if err == nil {
		t.Fatal("expected error for invalid range")
	}
}

func TestStatsService_CalculateStats_WithVariedData(t *testing.T) {
	mockAPI := &mock.MockLotteryAPI{}
	svc := NewStatsService(mockAPI)

	draws := make([]model.Draw, 100)
	for i := 0; i < 100; i++ {
		draws[i] = model.Draw{
			Issue:        "24001",
			FrontNumbers: [5]int{1, 2, 3, 4, 5},
			BackNumbers:  [2]int{6, 7},
		}
	}

	stats, err := svc.CalculateStats(context.Background(), draws, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stats.FrontHot) == 0 {
		t.Errorf("expected front hot numbers")
	}
	if len(stats.FrontCold) == 0 {
		t.Errorf("expected front cold numbers")
	}
}

func TestSplitHotCold(t *testing.T) {
	freqs := make([]model.NumberFrequency, 35)
	for i := range freqs {
		freqs[i] = model.NumberFrequency{Number: i + 1, Count: 35 - i}
	}

	// Need to sort first
	sortFrequencies(freqs)

	hot, warm, cold := splitHotCold(freqs)
	if len(hot) == 0 {
		t.Errorf("expected non-empty hot")
	}
	if len(warm) == 0 {
		t.Errorf("expected non-empty warm")
	}
	if len(cold) == 0 {
		t.Errorf("expected non-empty cold")
	}

	total := len(hot) + len(warm) + len(cold)
	if total != 35 {
		t.Errorf("expected total 35, got %d", total)
	}
}

func TestBuildFrequencies(t *testing.T) {
	freqMap := map[int]int{1: 5, 2: 3, 3: 8}
	lastSeen := map[int]int{1: 10, 2: 5, 3: 20}

	result := buildFrequencies(1, 3, freqMap, lastSeen, 30)
	if len(result) != 3 {
		t.Errorf("expected 3 frequencies, got %d", len(result))
	}
}
