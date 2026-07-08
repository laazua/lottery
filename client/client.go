// Package client 提供外部数据源访问抽象层，定义 LotteryAPI 接口。
package client

import (
	"context"
	"net/http"
	"time"

	"github.com/user/lottery/model"
)

// LotteryAPI 定义大乐透数据获取接口。
// 所有外部数据源实现必须满足此接口。
type LotteryAPI interface {
	// FetchDraws 拉取开奖数据列表。
	// opts 支持 WithBaseURL、WithPageSize、WithHTTPClient 等选项。
	FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error)

	// FetchDrawByPeriod 按期号查询单期开奖数据。
	FetchDrawByPeriod(ctx context.Context, period string) (*model.Draw, error)
}

// Option 是客户端选项的函数式接口。
type Option func(*options)

// options 包含客户端的所有配置项。
type options struct {
	baseURL    string
	pageSize   int
	httpClient HTTPDoer
	timeout    time.Duration
}

// HTTPDoer 是 HTTP 客户端的抽象接口。
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
