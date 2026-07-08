package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/user/lottery/internal/config"
	"github.com/user/lottery/internal/errors"
	"github.com/user/lottery/model"
)

// CWLClient 是 cwl.gov.cn 数据源的实现。
type CWLClient struct {
	opts    *options
	client  *http.Client
	baseURL *url.URL
}

// NewCWLClient 创建基于 config.APIBaseURL 的客户端。
func NewCWLClient(opts ...Option) *CWLClient {
	o := newOptions()
	for _, opt := range opts {
		opt(o)
	}

	baseURL, _ := url.Parse(o.baseURL)
	return &CWLClient{
		opts:    o,
		client:  &http.Client{Timeout: o.timeout},
		baseURL: baseURL,
	}
}

// FetchDraws 从配置的 API 拉取开奖数据列表。
// 当 config.DataSource 为 "mock" 时返回模拟数据。
func (c *CWLClient) FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error) {
	// 配置为 mock 模式时，直接返回模拟数据
	if config.DataSource == "mock" {
		slog.Info("使用模拟数据源")
		return NewMockClient().FetchDraws(ctx, opts...)
	}

	o := *c.opts
	for _, opt := range opts {
		opt(&o)
	}

	u := *c.baseURL
	// 尝试拼接 API 路径（不同数据源路径不同）
	u = *u.ResolveReference(&url.URL{Path: "/cwl_admin/front/findDraw"})
	q := u.Query()
	q.Set("name", "dlt")
	q.Set("issueCount", fmt.Sprintf("%d", o.pageSize))
	u.RawQuery = q.Encode()

	slog.Info("开始拉取开奖数据",
		"url", u.String(),
		"pageSize", o.pageSize,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		slog.Warn("API 不可用，回退到模拟数据", "error", err)
		return NewMockClient().FetchDraws(ctx, opts...)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		slog.Warn("API 请求失败，回退到模拟数据", "error", err)
		return NewMockClient().FetchDraws(ctx, opts...)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("API 返回非 200，回退到模拟数据", "status", resp.StatusCode)
		return NewMockClient().FetchDraws(ctx, opts...)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: 读取响应失败", errors.ErrParseResponse)
	}

	draws, err := parseDrawResponse(body)
	if err != nil {
		slog.Warn("API 响应解析失败，回退到模拟数据", "error", err)
		return NewMockClient().FetchDraws(ctx, opts...)
	}

	slog.Info("拉取开奖数据成功", "count", len(draws))
	return draws, nil
}

// FetchDrawByPeriod 按期号查询单期开奖数据。
func (c *CWLClient) FetchDrawByPeriod(ctx context.Context, period string) (*model.Draw, error) {
	draws, err := c.FetchDraws(ctx, WithPageSize(1))
	if err != nil {
		return nil, err
	}
	if len(draws) == 0 {
		return nil, errors.ErrEmptyResponse
	}
	return &draws[0], nil
}
