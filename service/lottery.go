// Package service 提供大乐透核心业务逻辑：数据拉取、统计分析、推荐算法。
package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/user/lottery/client"
	"github.com/user/lottery/internal/errors"
	"github.com/user/lottery/model"
)

// LotteryService 处理开奖数据的拉取与组装。
type LotteryService struct {
	api client.LotteryAPI
}

// NewLotteryService 创建 LotteryService 实例，依赖接口而非实现。
func NewLotteryService(api client.LotteryAPI) *LotteryService {
	return &LotteryService{api: api}
}

// FetchDraws 拉取开奖数据，支持分页。
func (s *LotteryService) FetchDraws(ctx context.Context, pageSize int) ([]model.Draw, error) {
	if pageSize <= 0 {
		return nil, fmt.Errorf("%w: pageSize 必须大于 0", errors.ErrInvalidParams)
	}

	draws, err := s.api.FetchDraws(ctx, client.WithPageSize(pageSize))
	if err != nil {
		slog.Error("拉取开奖数据失败", "pageSize", pageSize, "error", err)
		return nil, fmt.Errorf("拉取开奖数据: %w", err)
	}

	slog.Info("拉取开奖数据成功", "count", len(draws))
	return draws, nil
}

// FetchDrawsPage 分页拉取开奖数据，返回包含总记录数的分页结果。
func (s *LotteryService) FetchDrawsPage(ctx context.Context, pageNo, pageSize int) (*model.DrawsPage, error) {
	if pageNo <= 0 {
		return nil, fmt.Errorf("%w: pageNo 必须大于 0", errors.ErrInvalidParams)
	}
	if pageSize <= 0 {
		return nil, fmt.Errorf("%w: pageSize 必须大于 0", errors.ErrInvalidParams)
	}

	page, err := s.api.FetchDrawsPage(ctx, pageNo, pageSize)
	if err != nil {
		slog.Error("分页拉取开奖数据失败", "pageNo", pageNo, "pageSize", pageSize, "error", err)
		return nil, fmt.Errorf("分页拉取开奖数据: %w", err)
	}

	slog.Info("分页拉取开奖数据成功", "count", len(page.Draws), "total", page.Total, "pageNo", pageNo)
	return page, nil
}

// FetchDrawByPeriod 按期号查询单期开奖数据。
func (s *LotteryService) FetchDrawByPeriod(ctx context.Context, period string) (*model.Draw, error) {
	draw, err := s.api.FetchDrawByPeriod(ctx, period)
	if err != nil {
		slog.Error("按期号查询开奖数据失败", "period", period, "error", err)
		return nil, fmt.Errorf("查询 %s 期开奖数据: %w", period, err)
	}
	return draw, nil
}
