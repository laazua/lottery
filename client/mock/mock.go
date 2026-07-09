// Package mock 提供 LotteryAPI 的 Mock 实现，用于 service 层测试。
package mock

import (
	"context"

	"github.com/user/lottery/client"
	"github.com/user/lottery/model"
)

// MockLotteryAPI 是 LotteryAPI 的模拟实现。
type MockLotteryAPI struct {
	FetchDrawsFunc        func(ctx context.Context, opts ...client.Option) ([]model.Draw, error)
	FetchDrawByPeriodFunc func(ctx context.Context, period string) (*model.Draw, error)
	FetchDrawsPageFunc    func(ctx context.Context, pageNo, pageSize int) (*model.DrawsPage, error)
}

// FetchDraws 调用 Mock 函数或返回空结果。
func (m *MockLotteryAPI) FetchDraws(ctx context.Context, opts ...client.Option) ([]model.Draw, error) {
	if m.FetchDrawsFunc != nil {
		return m.FetchDrawsFunc(ctx, opts...)
	}
	return []model.Draw{}, nil
}

// FetchDrawByPeriod 调用 Mock 函数或返回空结果。
func (m *MockLotteryAPI) FetchDrawByPeriod(ctx context.Context, period string) (*model.Draw, error) {
	if m.FetchDrawByPeriodFunc != nil {
		return m.FetchDrawByPeriodFunc(ctx, period)
	}
	return nil, nil
}

// FetchDrawsPage 调用 Mock 函数或返回空结果。
func (m *MockLotteryAPI) FetchDrawsPage(ctx context.Context, pageNo, pageSize int) (*model.DrawsPage, error) {
	if m.FetchDrawsPageFunc != nil {
		return m.FetchDrawsPageFunc(ctx, pageNo, pageSize)
	}
	return &model.DrawsPage{Page: pageNo, PageSize: pageSize}, nil
}
