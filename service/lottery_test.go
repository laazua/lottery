package service

import (
	"context"
	"testing"
	"time"

	"github.com/user/lottery/client"
	"github.com/user/lottery/client/mock"
	"github.com/user/lottery/model"
)

func generateTestDraws(n int) []model.Draw {
	draws := make([]model.Draw, n)
	for i := 0; i < n; i++ {
		draws[i] = model.Draw{
			Issue:        "24001",
			DrawTime:     time.Now(),
			FrontNumbers: [5]int{1, 2, 3, 4, 5},
			BackNumbers:  [2]int{6, 7},
			SaleAmount:   100000000,
			PoolAmount:   500000000,
		}
	}
	return draws
}

func TestLotteryService_FetchDraws_Success(t *testing.T) {
	mockAPI := &mock.MockLotteryAPI{
		FetchDrawsFunc: func(ctx context.Context, opts ...client.Option) ([]model.Draw, error) {
			return []model.Draw{
				{Issue: "24180", FrontNumbers: [5]int{5, 12, 18, 23, 31}, BackNumbers: [2]int{7, 11}},
			}, nil
		},
	}
	svc := NewLotteryService(mockAPI)
	draws, err := svc.FetchDraws(context.Background(), 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(draws) != 1 {
		t.Errorf("expected 1 draw, got %d", len(draws))
	}
}

func TestLotteryService_FetchDraws_InvalidPageSize(t *testing.T) {
	mockAPI := &mock.MockLotteryAPI{}
	svc := NewLotteryService(mockAPI)
	_, err := svc.FetchDraws(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for invalid page size")
	}
}

func TestLotteryService_FetchDrawByPeriod(t *testing.T) {
	mockAPI := &mock.MockLotteryAPI{
		FetchDrawByPeriodFunc: func(ctx context.Context, period string) (*model.Draw, error) {
			return &model.Draw{Issue: period, FrontNumbers: [5]int{1, 2, 3, 4, 5}, BackNumbers: [2]int{6, 7}}, nil
		},
	}
	svc := NewLotteryService(mockAPI)
	draw, err := svc.FetchDrawByPeriod(context.Background(), "24180")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draw.Issue != "24180" {
		t.Errorf("expected issue 24180, got %s", draw.Issue)
	}
}
