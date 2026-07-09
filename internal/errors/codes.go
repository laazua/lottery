package errors

// ─── 通用错误 (E0) ───
var (
	ErrUnknown       = NewError("E0GN001", "未知错误")
	ErrInvalidParams = NewError("E0GN002", "参数校验失败")
	ErrUnsupported   = NewError("E0GN003", "不支持的请求")
)

// ─── 网络层错误 (E1) ───
var (
	ErrNetworkUnreachable = NewError("E1NT001", "网络不可达")
	ErrRequestTimeout     = NewError("E1NT002", "请求超时")
	ErrRateLimited        = NewError("E1NT003", "触发限流")
	ErrServerError        = NewError("E1NT004", "服务端错误")
	ErrTooManyRetries     = NewError("E1NT005", "超过最大重试次数")
)

// ─── 数据层错误 (E2) ───
var (
	ErrParseResponse      = NewError("E2DP001", "响应解析失败")
	ErrUnexpectedField    = NewError("E2DP002", "字段值异常")
	ErrEmptyResponse      = NewError("E2DP003", "响应数据为空")
	ErrAPIResponse        = NewError("E2DP004", "API 返回失败状态")
	ErrServiceUnavailable = NewError("E2SV001", "服务不可用")
)

// ─── 业务层错误 (E3) ───
var (
	ErrInsufficientDraws      = NewError("E3ST001", "统计数据不足")
	ErrInvalidStatsRange      = NewError("E3ST002", "统计范围参数无效")
	ErrNoValidRecommendation  = NewError("E3RC001", "无法生成有效推荐")
	ErrRecommendationDisabled = NewError("E3RC002", "推荐功能暂不可用")
)
