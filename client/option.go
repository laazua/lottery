package client

import (
	"net/http"
	"time"

	"github.com/user/lottery/internal/config"
	"github.com/user/lottery/internal/httputil"
)

// WithBaseURL 设置 API 基础 URL。
func WithBaseURL(url string) Option {
	return func(o *options) {
		o.baseURL = url
	}
}

// WithPageSize 设置每页期数。
func WithPageSize(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.pageSize = n
		}
	}
}

// WithHTTPClient 设置自定义 HTTP 客户端。
func WithHTTPClient(c HTTPDoer) Option {
	return func(o *options) {
		o.httpClient = c
	}
}

// WithTimeout 设置请求超时时间。
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.timeout = d
		}
	}
}

// newOptions 创建默认选项（从 config 包读取默认值）。
func newOptions() *options {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	if c := httputil.NewHTTPClient(10 * time.Second); c != nil {
		httpClient = c
	}
	return &options{
		baseURL:    config.APIBaseURL,
		pageSize:   20,
		httpClient: httpClient,
		timeout:    10 * time.Second,
	}
}
