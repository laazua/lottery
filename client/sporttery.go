// Package client 提供外部数据源访问抽象层。
package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/user/lottery/internal/errors"
	"github.com/user/lottery/model"
)

// sportteryDefaultBaseURL 是体彩 API 的默认基础 URL。
// 作为硬编码兜底，不受构建时 ldflags 覆盖的影响。
const sportteryDefaultBaseURL = "https://webapi.sporttery.cn"

// SportteryClient 是 webapi.sporttery.cn（中国体育彩票）数据源的实现。
type SportteryClient struct {
	opts    *options
	client  *http.Client
	baseURL *url.URL
}

// NewSportteryClient 创建基于配置的体彩客户端。
// 如果 config.APIBaseURL 为空或无效，使用硬编码默认值。
func NewSportteryClient(opts ...Option) *SportteryClient {
	o := newOptions()
	for _, opt := range opts {
		opt(o)
	}

	baseURL := resolveBaseURL(o.baseURL)
	return &SportteryClient{
		opts:    o,
		client:  &http.Client{Timeout: o.timeout},
		baseURL: baseURL,
	}
}

// resolveBaseURL 解析并校验 API 基础 URL。
// 如果传入的 URL 为空或缺少协议，回退到默认值。
func resolveBaseURL(rawURL string) *url.URL {
	if rawURL == "" {
		rawURL = sportteryDefaultBaseURL
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" {
		u, _ = url.Parse(sportteryDefaultBaseURL)
	}
	return u
}

// buildDrawsURL 拼接获取开奖数据的完整 URL。
func (c *SportteryClient) buildDrawsURL(pageSize, pageNo int) string {
	u := *c.baseURL
	u = *u.ResolveReference(&url.URL{Path: "/gateway/lottery/getHistoryPageListV1.qry"})
	q := u.Query()
	q.Set("gameNo", "85")    // 85 = 大乐透
	q.Set("provinceId", "0") // 0 = 全国
	q.Set("pageSize", strconv.Itoa(pageSize))
	q.Set("isVerify", "1")
	q.Set("pageNo", strconv.Itoa(pageNo))
	u.RawQuery = q.Encode()
	return u.String()
}

// newDrawsRequest 构建开奖数据 HTTP 请求。
func newDrawsRequest(ctx context.Context, urlStr string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; Xiaomi) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.lottery.gov.cn/")
	req.Header.Set("Origin", "https://www.lottery.gov.cn")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	return req, nil
}

// doDrawsRequest 执行 HTTP 请求并读取响应体。
func (c *SportteryClient) doDrawsRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("体彩 API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: 体彩 API 返回状态码 %d", errors.ErrAPIResponse, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: 读取响应失败", errors.ErrParseResponse)
	}
	return body, nil
}

// FetchDraws 从体彩官方 API 拉取大乐透历史开奖数据。
// 获取失败时直接返回 error，不会回落至模拟数据。
func (c *SportteryClient) FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error) {
	o := *c.opts
	for _, opt := range opts {
		opt(&o)
	}

	urlStr := c.buildDrawsURL(o.pageSize, o.pageNo)
	slog.Info("开始拉取体彩开奖数据", "url", urlStr, "pageSize", o.pageSize)

	req, err := newDrawsRequest(ctx, urlStr)
	if err != nil {
		return nil, fmt.Errorf("构建 API 请求失败: %w", err)
	}

	body, err := c.doDrawsRequest(req)
	if err != nil {
		return nil, err
	}

	draws, _, err := parseSportteryResponse(body)
	if err != nil {
		return nil, fmt.Errorf("解析体彩 API 响应失败: %w", err)
	}

	slog.Info("拉取体彩开奖数据成功", "count", len(draws))
	return draws, nil
}

// FetchDrawsPage 分页拉取大乐透历史开奖数据。
// 返回包含总记录数的分页结果。
func (c *SportteryClient) FetchDrawsPage(ctx context.Context, pageNo, pageSize int) (*model.DrawsPage, error) {
	urlStr := c.buildDrawsURL(pageSize, pageNo)
	slog.Info("开始拉取体彩开奖数据（分页）",
		"url", urlStr, "pageNo", pageNo, "pageSize", pageSize,
	)

	req, err := newDrawsRequest(ctx, urlStr)
	if err != nil {
		return nil, fmt.Errorf("构建 API 请求失败: %w", err)
	}

	body, err := c.doDrawsRequest(req)
	if err != nil {
		return nil, err
	}

	draws, total, err := parseSportteryResponse(body)
	if err != nil {
		return nil, fmt.Errorf("解析体彩 API 响应失败: %w", err)
	}

	slog.Info("拉取体彩开奖数据成功（分页）",
		"count", len(draws), "total", total, "pageNo", pageNo,
	)
	return &model.DrawsPage{
		Draws:    draws,
		Total:    total,
		Page:     pageNo,
		PageSize: pageSize,
	}, nil
}

// FetchDrawByPeriod 按期号查询单期开奖数据。
func (c *SportteryClient) FetchDrawByPeriod(ctx context.Context, period string) (*model.Draw, error) {
	page, err := c.FetchDrawsPage(ctx, 1, 1)
	if err != nil {
		return nil, err
	}
	if len(page.Draws) == 0 {
		return nil, errors.ErrEmptyResponse
	}
	return &page.Draws[0], nil
}
