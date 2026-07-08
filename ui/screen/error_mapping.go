package screen

import (
	"errors"

	lottErrors "github.com/user/lottery/internal/errors"
)

// ErrToUserMessage 将业务错误转换为用户友好的提示文案。
func ErrToUserMessage(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, lottErrors.ErrNetworkUnreachable):
		return "网络连接失败，请检查网络设置"
	case errors.Is(err, lottErrors.ErrRequestTimeout):
		return "请求超时，请稍后重试"
	case errors.Is(err, lottErrors.ErrRateLimited):
		return "查询过于频繁，请稍后再试"
	case errors.Is(err, lottErrors.ErrServerError):
		return "数据源暂时不可用"
	case errors.Is(err, lottErrors.ErrEmptyResponse):
		return "暂无相关开奖数据"
	case errors.Is(err, lottErrors.ErrParseResponse):
		return "数据格式异常"
	case errors.Is(err, lottErrors.ErrInsufficientDraws):
		return "历史数据不足，暂无法统计"
	case errors.Is(err, lottErrors.ErrInvalidStatsRange):
		return "请选择有效的统计期数"
	case errors.Is(err, lottErrors.ErrNoValidRecommendation):
		return "数据不足，无法生成推荐"
	default:
		return "出了点问题，请稍后重试"
	}
}
