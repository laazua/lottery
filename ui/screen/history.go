// Package screen 提供各业务屏的布局编排逻辑。
package screen

import (
	"context"
	"fmt"
	"image"
	"log/slog"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/user/lottery/model"
	"github.com/user/lottery/ui/theme"
	lotwidget "github.com/user/lottery/ui/widget"
)

// HistoryState 维护历史开奖屏的状态（含所有 gioui 交互组件）。
type HistoryState struct {
	Draws   []model.Draw
	Loading bool
	Loaded  bool // 是否已触发过首次加载
	Error   error

	// gioui 交互组件（必须持久化！）
	RefreshBtn widget.Clickable
	List       layout.List
	searchBtn  widget.Clickable
}

// HistoryLayout 渲染历史开奖查询页面。
func HistoryLayout(gtx layout.Context, th *theme.Theme, state *HistoryState, svc *Services, drawsCache *[]model.Draw) layout.Dimensions {
	// ═══ ① 事件检测（布局代码之前）═══
	if !state.Loaded && !state.Loading {
		state.Loading = true
		state.Loaded = true
		go fetchDrawsAsync(state, svc)
	}

	if state.RefreshBtn.Clicked(gtx) {
		state.Loading = true
		state.Loaded = false // 允许重新加载
		go fetchDrawsAsync(state, svc)
	}

	// ═══ ② 页面布局 ═══
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// 顶部标题栏
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return historyHeader(gtx, th, state)
		}),
		// 列表内容
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return historyContent(gtx, th, state)
		}),
	)
}

// historyHeader 渲染顶部标题栏（含刷新按钮）。
func historyHeader(gtx layout.Context, th *theme.Theme, state *HistoryState) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Small, Bottom: th.Spacing.Small,
		Left: th.Spacing.Medium, Right: th.Spacing.Small,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Horizontal,
		}.Layout(gtx,
			// 标题
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				title := "开奖查询"
				lbl := material.Label(th.Theme, unit.Sp(20), title)
				lbl.Font.Weight = font.Bold
				lbl.Color = th.Colors.OnSurface
				return lbl.Layout(gtx)
			}),
			// 刷新按钮（使用胶囊形填充按钮）
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := lotwidget.FilledBtn(th, "刷新", &state.RefreshBtn)
				return btn.Layout(gtx)
			}),
		)
	})
}

// historyContent 渲染列表内容或空状态/加载状态。
func historyContent(gtx layout.Context, th *theme.Theme, state *HistoryState) layout.Dimensions {
	if state.Loading && len(state.Draws) == 0 {
		return lotwidget.LoadingSkeleton(gtx, th)
	}

	if state.Error != nil && len(state.Draws) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(16), "数据加载失败")
					lbl.Color = th.Colors.Error
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(12), "请检查网络后点击刷新")
					lbl.Color = th.Colors.Disabled
					return lbl.Layout(gtx)
				}),
			)
		})
	}

	if len(state.Draws) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.Theme, unit.Sp(16), "暂无开奖数据")
			lbl.Color = th.Colors.Disabled
			return lbl.Layout(gtx)
		})
	}

	// 列表渲染
	return state.List.Layout(gtx, len(state.Draws), func(gtx layout.Context, index int) layout.Dimensions {
		return drawCard(gtx, th, state.Draws[index])
	})
}

// drawCard 渲染单个开奖卡片（MD3 风格）。
func drawCard(gtx layout.Context, th *theme.Theme, draw model.Draw) layout.Dimensions {
	// 内边距
	return layout.Inset{
		Top: th.Spacing.XXSmall, Bottom: th.Spacing.XXSmall,
		Left: th.Spacing.Small, Right: th.Spacing.Small,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 白色圆角卡片
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				// 卡片背景
				r := gtx.Dp(th.Shape.Medium)
				defer clip.RRect{
					Rect: image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
					NE: r, NW: r, SE: r, SW: r,
				}.Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, th.Colors.Surface)

				return layout.UniformInset(th.Spacing.Medium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						// 期号 + 日期
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{
								Axis: layout.Horizontal,
							}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(th.Theme, unit.Sp(15), fmt.Sprintf("第 %s 期", draw.Issue))
									lbl.Font.Weight = font.Bold
									lbl.Color = th.Colors.OnSurface
									return lbl.Layout(gtx)
								}),
							)
						}),
						// 号码球行
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top: th.Spacing.Small, Bottom: th.Spacing.Small,
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{
									Axis: layout.Horizontal,
								}.Layout(gtx, drawBalls(th, draw)...)
							})
						}),
						// 底部信息
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if draw.SaleAmount > 0 || draw.PoolAmount > 0 {
								info := fmt.Sprintf("销售额: %.1f亿  奖池: %.1f亿",
									float64(draw.SaleAmount)/1e8,
									float64(draw.PoolAmount)/1e8,
								)
								lbl := material.Label(th.Theme, unit.Sp(11), info)
								lbl.Color = th.Colors.Disabled
								return lbl.Layout(gtx)
							}
							return layout.Dimensions{}
						}),
					)
				})
			}),
		)
	})
}

// drawBalls 构建号码球列表（前区5蓝 + 分隔 + 后区2红）。
func drawBalls(th *theme.Theme, draw model.Draw) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, 7)
	for _, n := range draw.FrontNumbers {
		n := n
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return lotwidget.Ball(gtx, th, n, lotwidget.BallNormal, th.BallSizes.Small)
		}))
	}
	// 分隔符
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Left: th.Spacing.XSmall, Right: th.Spacing.XSmall,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.Theme, unit.Sp(14), "+")
			lbl.Color = th.Colors.Disabled
			return lbl.Layout(gtx)
		})
	}))
	for _, n := range draw.BackNumbers {
		n := n
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return lotwidget.Ball(gtx, th, n, lotwidget.BallHot, th.BallSizes.Small)
		}))
	}
	return children
}

// fetchDrawsAsync 在 goroutine 中异步拉取开奖数据。
func fetchDrawsAsync(state *HistoryState, svc *Services) {
	defer func() {
		state.Loading = false
		if svc.Invalidate != nil {
			svc.Invalidate()
		}
	}()

	draws, err := svc.Lottery.FetchDraws(context.Background(), 20)
	if err != nil {
		slog.Error("拉取开奖数据失败", "error", err)
		state.Error = err
		return
	}

	slog.Info("拉取开奖数据成功", "count", len(draws))
	state.Draws = draws
	state.Error = nil
}
