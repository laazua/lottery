// Package httputil 提供 HTTP 重试、超时、限流等通用工具。
package httputil

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/user/lottery/internal/errors"
)

// DefaultTimeout 默认 HTTP 请求超时时间。
const DefaultTimeout = 10 * time.Second

// NewHTTPClient 创建带默认超时的 HTTP 客户端。
func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &http.Client{Timeout: timeout}
}

// RetryWithBackoff 使用指数退避策略执行重试。
// maxRetries 为最大重试次数（不包括首次尝试），
// fn 为需要重试的操作，返回 error 表示需要重试。
// 如果 ctx 被取消，立即返回 ctx.Err() 不重试。
func RetryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := (1 << (attempt - 1)) * time.Second
			jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
			wait := backoff + jitter

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		return nil
	}

	return fmt.Errorf("%w: 已重试 %d 次, 最后错误: %v", errors.ErrTooManyRetries, maxRetries, lastErr)
}
