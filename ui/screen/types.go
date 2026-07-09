// Package screen 提供导航路由和共享类型定义。
package screen

import "github.com/user/lottery/service"

// ScreenID 标识当前显示的屏幕。
type ScreenID int

const (
	ScreenHistory ScreenID = iota
	ScreenStats
	ScreenRecommend
	ScreenRandom
)

// Route 维护导航状态。
type Route struct {
	Current ScreenID
}

// Services 聚合所有业务服务和 UI 回调，供各屏使用。
type Services struct {
	Lottery    *service.LotteryService
	Stats      *service.StatsService
	Recommend  *service.RecommendService
	Invalidate func() // 触发 UI 重绘，由 app.Run 传入 w.Invalidate
}
