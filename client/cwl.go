// Package client 提供外部数据源访问抽象层。
package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/user/lottery/internal/errors"
	"github.com/user/lottery/model"
)

// CWLClient 是 cwl.gov.cn 数据源的实现。
type CWLClient struct {
	opts    *options
	client  *http.Client
	baseURL *url.URL
}

// NewCWLClient 创建基于配置的客户端。
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
// 获取失败时直接返回 error，不会回落至模拟数据。
func (c *CWLClient) FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error) {
	o := *c.opts
	for _, opt := range opts {
		opt(&o)
	}

	u := *c.baseURL
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
		return nil, fmt.Errorf("构建 API 请求失败: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: API 返回状态码 %d", errors.ErrAPIResponse, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: 读取响应失败", errors.ErrParseResponse)
	}

	draws, err := parseDrawResponse(body)
	if err != nil {
		return nil, fmt.Errorf("解析 API 响应失败: %w", err)
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

// FetchDrawsPage 分页拉取开奖数据。
// CWL 接口不直接支持分页，直接返回所有数据。
func (c *CWLClient) FetchDrawsPage(ctx context.Context, pageNo, pageSize int) (*model.DrawsPage, error) {
	draws, err := c.FetchDraws(ctx, WithPageSize(pageSize))
	if err != nil {
		return nil, err
	}
	return &model.DrawsPage{
		Draws:    draws,
		Total:    len(draws),
		Page:     pageNo,
		PageSize: pageSize,
	}, nil
}
