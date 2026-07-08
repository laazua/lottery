package client

import (
	"context"
	"fmt"
	"time"

	"github.com/user/lottery/model"
)

// drawRow 用于构建模拟数据的单期开奖。
type drawRow struct {
	issue         string
	drawTime      string
	frontNums     [5]int
	backNums      [2]int
	saleAmount    int64
	poolAmount    int64
}

// mockDraws 预置 30 期大乐透模拟数据，基于真实开奖规律生成。
var mockDraws = func() []drawRow {
	rows := make([]drawRow, 0, 30)
	// 生成近 30 期模拟数据，期号递减、号码随机但接近真实分布
	for i := 0; i < 30; i++ {
		issueNum := 24180 - i
		date := time.Date(2026, 7, 8, 21, 20, 0, 0, time.Local).AddDate(0, 0, -3*i)
		rows = append(rows, drawRow{
			issue:    fmt.Sprintf("%d", issueNum),
			drawTime: date.Format("2006-01-02 15:04:05"),
			frontNums: [5]int{
				frontPool[(i*7+0)%35],
				frontPool[(i*7+1)%35],
				frontPool[(i*7+2)%35],
				frontPool[(i*7+3)%35],
				frontPool[(i*7+4)%35],
			},
			backNums: [2]int{
				backPool[(i*3+0)%12],
				backPool[(i*3+1)%12],
			},
			saleAmount: 300_000_000 + int64(i)*10_000_000,
			poolAmount: 900_000_000 + int64(i)*5_000_000,
		})
	}
	return rows
}()

// frontPool 前区号码池（1-35），排列经过随机打散模拟真实分布。
var frontPool = [35]int{
	7, 12, 18, 23, 25, 31, 33, 5, 14, 21,
	1, 9, 17, 20, 27, 29, 34, 3, 11, 15,
	22, 26, 30, 35, 4, 8, 13, 16, 19, 24,
	28, 32, 2, 6, 10,
}

// backPool 后区号码池（1-12）。
var backPool = [12]int{
	3, 7, 11, 2, 8, 1, 9, 4, 10, 5, 12, 6,
}

// MockClient 返回模拟开奖数据的客户端实现。
type MockClient struct{}

// NewMockClient 创建模拟数据客户端。
func NewMockClient() *MockClient {
	return &MockClient{}
}

// FetchDraws 返回预置的模拟开奖数据。
func (m *MockClient) FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error) {
	o := newOptions()
	for _, opt := range opts {
		opt(o)
	}

	count := o.pageSize
	if count > len(mockDraws) {
		count = len(mockDraws)
	}

	draws := make([]model.Draw, 0, count)
	for i := 0; i < count; i++ {
		row := mockDraws[i]
		t, err := time.Parse("2006-01-02 15:04:05", row.drawTime)
		if err != nil {
			continue
		}
		draws = append(draws, model.Draw{
			Issue:    row.issue,
			DrawTime: t,
			FrontNumbers: row.frontNums,
			BackNumbers:  row.backNums,
			SaleAmount:   row.saleAmount,
			PoolAmount:   row.poolAmount,
		})
	}
	return draws, nil
}

// FetchDrawByPeriod 按期号查询单期模拟数据。
func (m *MockClient) FetchDrawByPeriod(ctx context.Context, period string) (*model.Draw, error) {
	for _, row := range mockDraws {
		if row.issue == period {
			t, _ := time.Parse("2006-01-02 15:04:05", row.drawTime)
			return &model.Draw{
				Issue:    row.issue,
				DrawTime: t,
				FrontNumbers: row.frontNums,
				BackNumbers:  row.backNums,
				SaleAmount:   row.saleAmount,
				PoolAmount:   row.poolAmount,
			}, nil
		}
	}
	// 找不到返回全部数据中的第一期
	draws, _ := m.FetchDraws(ctx)
	if len(draws) > 0 {
		return &draws[0], nil
	}
	return nil, fmt.Errorf("模拟数据为空")
}
