# 05-errors-code.md — 错误码规范

## 1. 错误码结构

### 1.1 编码规则

错误码采用 **层级编码体系**，格式为：

```
E[层级][模块][序号]
```

| 位置 | 说明 | 取值范围 |
|------|------|---------|
| E | 固定前缀，表示 Error | — |
| 层级 | 0=通用, 1=网络, 2=数据, 3=业务 | 0-3 |
| 模块 | 2位字母缩写 | 见下方模块表 |
| 序号 | 3位数字编号 | 001-999 |

### 1.2 模块编码表

| 编码 | 模块 | 所属层级 |
|------|------|---------|
| GN | 通用 (General) | 0 |
| NT | 网络 (Network) | 1 |
| DP | 数据解析 (Data Parse) | 2 |
| SV | 服务 (Service) | 2 |
| ST | 统计 (Statistics) | 3 |
| RC | 推荐 (Recommendation) | 3 |

## 2. 通用错误码表

### 层级 0：通用错误 (E0)

| 错误码 | 错误常量 | HTTP类比 | 说明 |
|--------|---------|---------|------|
| E0GN001 | `ErrUnknown` | 500 | 未知错误 |
| E0GN002 | `ErrInvalidParams` | 400 | 参数校验失败 |
| E0GN003 | `ErrUnsupported` | — | 不支持的请求 |

### 层级 1：网络层错误 (E1)

| 错误码 | 错误常量 | HTTP类比 | 说明 | 用户提示 |
|--------|---------|---------|------|---------|
| E1NT001 | `ErrNetworkUnreachable` | — | 网络不可达 | "网络连接失败，请检查网络设置" |
| E1NT002 | `ErrRequestTimeout` | 408 | 请求超时 | "请求超时，请稍后重试" |
| E1NT003 | `ErrRateLimited` | 429 | 触发限流 | "查询过于频繁，请稍后再试" |
| E1NT004 | `ErrServerError` | 500 | 服务端错误 | "数据源暂时不可用" |
| E1NT005 | `ErrTooManyRetries` | — | 超过最大重试次数 | "服务暂时不可用，请稍后重试" |

### 层级 2：数据层错误 (E2)

| 错误码 | 错误常量 | 说明 | 用户提示 |
|--------|---------|------|---------|
| E2DP001 | `ErrParseResponse` | 响应 JSON 解析失败 | "数据格式异常" |
| E2DP002 | `ErrUnexpectedField` | 响应字段值不在预期范围 | "数据内容异常" |
| E2DP003 | `ErrEmptyResponse` | 响应数据为空 | "暂无相关数据" |
| E2SV001 | `ErrServiceUnavailable` | 服务不可用 | "服务暂不可用" |

### 层级 3：业务层错误 (E3)

| 错误码 | 错误常量 | 说明 | 用户提示 |
|--------|---------|------|---------|
| E3ST001 | `ErrInsufficientDraws` | 统计数据不足（少于最低期数要求） | "历史数据不足，暂无法统计" |
| E3ST002 | `ErrInvalidStatsRange` | 统计范围参数非法 | "请选择有效的统计期数" |
| E3RC001 | `ErrNoValidRecommendation` | 无法生成有效推荐（数据不足） | "数据不足，无法生成推荐" |
| E3RC002 | `ErrRecommendationDisabled` | 推荐功能暂时不可用 | "推荐功能暂不可用" |

## 3. 错误码定义

### 3.1 Error 类型定义

```go
// internal/errors/error.go

package errors

import "fmt"

// Error 是带错误码的业务错误类型。
type Error struct {
    Code    string // 错误码，如 "E1NT001"
    Message string // 人类可读的错误描述
}

func (e *Error) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Is 支持 errors.Is 匹配错误码。
func (e *Error) Is(target error) bool {
    // 如果 target 也是 *Error，比较 Code
    var t *Error
    if ok := as(target, &t); ok {
        return e.Code == t.Code
    }
    return false
}

// NewError 创建带错误码的业务错误。
func NewError(code, msg string) *Error {
    return &Error{Code: code, Message: msg}
}
```

### 3.2 预定义错误码

```go
// internal/errors/codes.go

package errors

// ─── 通用错误 ───
var ErrUnknown       = NewError("E0GN001", "未知错误")
var ErrInvalidParams = NewError("E0GN002", "参数校验失败")

// ─── 网络错误 ───
var ErrNetworkUnreachable = NewError("E1NT001", "网络不可达")
var ErrRequestTimeout     = NewError("E1NT002", "请求超时")
var ErrRateLimited        = NewError("E1NT003", "触发限流")
var ErrServerError        = NewError("E1NT004", "服务端错误")
var ErrTooManyRetries     = NewError("E1NT005", "超过最大重试次数")

// ─── 数据层错误 ───
var ErrParseResponse      = NewError("E2DP001", "响应解析失败")
var ErrUnexpectedField    = NewError("E2DP002", "字段值异常")
var ErrEmptyResponse      = NewError("E2DP003", "响应数据为空")
var ErrServiceUnavailable = NewError("E2SV001", "服务不可用")

// ─── 业务错误 ───
var ErrInsufficientDraws        = NewError("E3ST001", "统计数据不足")
var ErrInvalidStatsRange        = NewError("E3ST002", "统计范围参数无效")
var ErrNoValidRecommendation    = NewError("E3RC001", "无法生成有效推荐")
var ErrRecommendationDisabled   = NewError("E3RC002", "推荐功能暂不可用")
```

## 4. 错误处理流程

### 4.1 完整链路

```
┌──────────┐     ┌──────────────┐     ┌──────────┐     ┌───────────┐
│ client/  │ ──▶ │  service/    │ ──▶ │  ui/     │ ──▶ │ 用户      │
│ 返回 err │     │ 包装 + 记录   │     │ 展示     │     │ 看到提示  │
└──────────┘     └──────────────┘     └──────────┘     └───────────┘
    │                  │                  │
    ▼                  ▼                  ▼
 原始错误           包装错误           slog.Error 记录
```

### 4.2 各层职责

| 层次 | 职责 |
|------|------|
| `client/` | 识别 HTTP 状态码，映射到对应错误码；附加请求上下文 |
| `service/` | 使用 `fmt.Errorf(%w)` 包装错误；调用 slog 记录日志；判定是否可重试 |
| `ui/screen/` | 提取错误码 → 映射用户提示文案 → Toast 展示；提供重试入口 |

### 4.3 错误 → UI 提示映射

```go
// ui/screen/error_mapping.go

func userMessage(err error) string {
    switch {
    case errors.Is(err, ErrNetworkUnreachable):
        return "网络连接失败，请检查网络设置"
    case errors.Is(err, ErrRequestTimeout):
        return "请求超时，请稍后重试"
    case errors.Is(err, ErrEmptyResponse):
        return "暂无相关开奖数据"
    case errors.Is(err, ErrInsufficientDraws):
        return "历史数据不足，暂时无法统计"
    default:
        return "出了点问题，请稍后重试"
    }
}
```

## 5. 错误日志规范

### 5.1 日志格式

```go
// ✅ 标准日志格式
slog.Error("拉取开奖数据失败",
    "error_code", "E1NT001",
    "period", "24180",
    "attempt", 3,
    "error", err,
)
```

### 5.2 日志级别与错误码对照

| 错误码范围 | 日志级别 | 说明 |
|-----------|---------|------|
| E0xxxx | Error | 未知错误需人工关注 |
| E1NT001 | Error | 网络不可达 |
| E1NT002 | Warn → Error | 首次重试 Warn，最终失败 Error |
| E1NT003 | Warn | 限流属暂时状态 |
| E2xxxx | Error | 数据异常需排查 |
| E3xxxx | Info/Warn | 业务正常状态，非系统问题 |

## 6. 测试要求

- 每个错误码至少有一个单元测试覆盖对应的场景
- `client/` 层 mock HTTP 响应测试各错误码映射
- `service/` 层测试错误包装是否正确传递了根因错误
- `ui/` 层测试错误码到用户提示的映射是否正确
