// Package ui 提供 gioui 窗口初始化、全局状态管理和应用生命周期适配。
package ui

import (
	"log/slog"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	giawidget "gioui.org/widget"

	"github.com/user/lottery/model"
	"github.com/user/lottery/service"
	"github.com/user/lottery/ui/screen"
	"github.com/user/lottery/ui/theme"
)

// AppState 包含所有屏幕的独立状态。
type AppState struct {
	History   screen.HistoryState
	Stats     screen.StatsState
	Recommend screen.RecommendState
}

// App 是应用的主控制器，管理窗口、状态和依赖。
type App struct {
	Window     *app.Window
	State      AppState
	Route      screen.Route
	Theme      *theme.Theme
	Services   screen.Services
	TabBtns    [3]giawidget.Clickable
	drawsCache []model.Draw
}

// NewApp 创建 App 实例并组装依赖。
func NewApp(lotterySvc *service.LotteryService, statsSvc *service.StatsService, recommSvc *service.RecommendService) *App {
	w := new(app.Window)
	w.Option(
		app.Title("大乐透助手"),
		app.Size(unit.Dp(400), unit.Dp(700)),
	)
	return &App{
		Window: w,
		Route:  screen.Route{Current: screen.ScreenHistory},
		Theme:  theme.NewTheme(),
		Services: screen.Services{
			Lottery:   lotterySvc,
			Stats:     statsSvc,
			Recommend: recommSvc,
		},
	}
}

// Run 启动应用的事件循环。
func (a *App) Run() error {
	var ops op.Ops

	// 注册 Invalidate 回调（需要在拿到 Window 后绑定）
	a.Services.Invalidate = a.Window.Invalidate

	slog.Info("大乐透助手事件循环已启动")

	for {
		e := a.Window.Event()
		switch evt := e.(type) {
		case app.DestroyEvent:
			return evt.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, evt)
			a.Layout(gtx)
			evt.Frame(gtx.Ops)
		}
	}
}

// Layout 编排当前屏幕的布局和底部导航。
func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	// ① 检测 Tab 点击事件
	for i := range a.TabBtns {
		if a.TabBtns[i].Clicked(gtx) {
			if screen.ScreenID(i) != a.Route.Current {
				a.Route.Current = screen.ScreenID(i)
				slog.Info("切换屏幕", "target", a.Route.Current)
			}
		}
	}

	// ② 整体布局：内容区 + 底部导航栏
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// 内容区域
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			switch a.Route.Current {
			case screen.ScreenHistory:
				return screen.HistoryLayout(gtx, a.Theme, &a.State.History, &a.Services, &a.drawsCache)
			case screen.ScreenStats:
				return screen.StatsLayout(gtx, a.Theme, &a.State.Stats, &a.Services, &a.drawsCache)
			case screen.ScreenRecommend:
				return screen.RecommendLayout(gtx, a.Theme, &a.State.Recommend, &a.Services, &a.drawsCache)
			default:
				return screen.HistoryLayout(gtx, a.Theme, &a.State.History, &a.Services, &a.drawsCache)
			}
		}),
		// 底部导航栏
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return screen.BottomNavLayout(gtx, a.Theme, a.Route.Current, &a.TabBtns)
		}),
	)
}
