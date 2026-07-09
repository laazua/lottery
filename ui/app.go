// Package ui 提供 gioui 窗口初始化、全局状态管理和应用生命周期适配。
package ui

import (
	"image"
	"log/slog"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	giawidget "gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/user/lottery/model"
	"github.com/user/lottery/service"
	"github.com/user/lottery/ui/screen"
	"github.com/user/lottery/ui/theme"
	lotwidget "github.com/user/lottery/ui/widget"
)

// AppState 包含所有屏幕的独立状态。
type AppState struct {
	History   screen.HistoryState
	Stats     screen.StatsState
	Recommend screen.RecommendState
	Random    screen.RandomState
}

// App 是应用的主控制器，管理窗口、状态和依赖。
type App struct {
	Window     *app.Window
	State      AppState
	Route      screen.Route
	Theme      *theme.Theme
	Services   screen.Services
	TabBtns    [4]giawidget.Clickable
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

// Layout 编排 Header、当前屏幕布局和底部导航。
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

	// ② 顶层：填充页面暖调米灰背景（基础层）
	paint.Fill(gtx.Ops, a.Theme.Colors.Bg)

	// ③ 整体布局：Header + 内容区 + 底部导航栏
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// 顶部蓝色 Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return header(gtx, a.Theme)
		}),
		// 内容区域
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			switch a.Route.Current {
			case screen.ScreenHistory:
				return screen.HistoryLayout(gtx, a.Theme, &a.State.History, &a.Services, &a.drawsCache)
			case screen.ScreenStats:
				return screen.StatsLayout(gtx, a.Theme, &a.State.Stats, &a.Services, &a.drawsCache)
			case screen.ScreenRecommend:
				return screen.RecommendLayout(gtx, a.Theme, &a.State.Recommend, &a.Services, &a.drawsCache)
			case screen.ScreenRandom:
				return screen.RandomLayout(gtx, a.Theme, &a.State.Random)
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

// header 渲染蓝色顶部标题栏（含标题和示例号码）。
func header(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	headerH := gtx.Dp(unit.Dp(130))
	defer clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, headerH)).Push(gtx.Ops).Pop()

	gtx.Constraints.Min.Y = headerH
	gtx.Constraints.Max.Y = headerH

	paint.Fill(gtx.Ops, th.Colors.Primary)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// 标题行
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left: unit.Dp(16),
				Top:  unit.Dp(24),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th.Theme, unit.Sp(20), "大乐透助手")
				lbl.Font.Weight = font.Bold
				lbl.Color = th.Colors.OnPrimary
				return lbl.Layout(gtx)
			})
		}),
		// 号码行（居中）
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(gtx, headerBallRow(th)...)
				})
			})
		}),
	)
}

// headerBallRow 渲染 header 中的示例号码球（前区5蓝 + 后区2红）。
func headerBallRow(th *theme.Theme) []layout.FlexChild {
	frontNums := []int{5, 12, 18, 23, 31}
	backNums := []int{7, 11}
	var children []layout.FlexChild

	for _, n := range frontNums {
		n := n
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Right: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return lotwidget.Ball(gtx, th, n, lotwidget.BallFront, unit.Dp(22))
			})
		}))
	}

	// 分隔线
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Left: unit.Dp(4), Right: unit.Dp(4),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.Theme, unit.Sp(16), "|")
			lbl.Color = th.Colors.OnPrimary
			return lbl.Layout(gtx)
		})
	}))

	for _, n := range backNums {
		n := n
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Right: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return lotwidget.Ball(gtx, th, n, lotwidget.BallBack, unit.Dp(22))
			})
		}))
	}
	return children
}
