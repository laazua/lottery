package screen

import (
	"context"
	"fmt"
	"image"
	"image/color"
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

// periodOption 统计期数范围选项。
type periodOption struct {
	Label string
	Value int
}

var periodOptions = []periodOption{
	{Label: "近20期", Value: 20},
	{Label: "近50期", Value: 50},
	{Label: "近100期", Value: 100},
}

// StatsState 维护冷热统计屏的状态。
type StatsState struct {
	Stats       *model.Statistics
	Loading     bool
	Loaded      bool
	Error       error
	PeriodRange int

	// gioui 交互组件（持久化）
	periodBtns [3]widget.Clickable
	List       layout.List
}

// StatsLayout 渲染冷热统计页面。
func StatsLayout(gtx layout.Context, th *theme.Theme, state *StatsState, svc *Services, drawsCache *[]model.Draw) layout.Dimensions {
	// ═══ ① 事件检测 ═══
	// 修复 gioui layout.List 零值默认 Horizontal 导致列表不可见的问题。
	if state.List.Axis != layout.Vertical {
		state.List.Axis = layout.Vertical
	}

	for i := range periodOptions {
		if state.periodBtns[i].Clicked(gtx) {
			state.PeriodRange = periodOptions[i].Value
			state.Loaded = false
			state.Loading = true
			go fetchStatsAsync(state, svc, drawsCache)
		}
	}

	if !state.Loaded && !state.Loading {
		state.PeriodRange = 20
		state.Loading = true
		go fetchStatsAsync(state, svc, drawsCache)
	}

	// ═══ ② 页面布局 ═══
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return statsHeader(gtx, th, state)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return statsContent(gtx, th, state)
		}),
	)
}

// statsHeader 渲染统计页标题和期数选择器。
func statsHeader(gtx layout.Context, th *theme.Theme, state *StatsState) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Medium, Bottom: th.Spacing.Small,
		Left: th.Spacing.Large, Right: th.Spacing.Large,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// 标题
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th.Theme, unit.Sp(18), "冷热统计")
				lbl.Font.Weight = font.Bold
				lbl.Color = th.Colors.OnSurface
				return lbl.Layout(gtx)
			}),
			// 期数选择器
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top: th.Spacing.XSmall,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(gtx, func() []layout.FlexChild {
						var children []layout.FlexChild
						for i, opt := range periodOptions {
							i, opt := i, opt
							children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{
									Right: th.Spacing.Small,
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									selected := state.PeriodRange == opt.Value
									var btn material.ButtonStyle
									if selected {
										btn = lotwidget.SmallFilledBtn(th, opt.Label, &state.periodBtns[i])
									} else {
										btn = lotwidget.OutlineBtn(th, opt.Label, &state.periodBtns[i])
									}
									return btn.Layout(gtx)
								})
							}))
						}
						return children
					}()...)
				})
			}),
		)
	})
}

// statsContent 渲染统计内容。
func statsContent(gtx layout.Context, th *theme.Theme, state *StatsState) layout.Dimensions {
	if state.Error != nil && state.Stats == nil {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(16), "统计数据加载失败")
					lbl.Color = th.Colors.Error
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(12), "请检查网络后点击期数按钮重试")
					lbl.Color = th.Colors.Disabled
					return lbl.Layout(gtx)
				}),
			)
		})
	}

	if state.Loading && state.Stats == nil {
		return lotwidget.LoadingSkeleton(gtx, th)
	}

	if state.Stats == nil {
		return lotwidget.LoadingIndicator(gtx, th, "加载统计数据中...")
	}

	// 滚动的统计内容
	return state.List.Layout(gtx, 3, func(gtx layout.Context, index int) layout.Dimensions {
		switch index {
		case 0:
			return numberSection(gtx, th, "前区号码热度", state.Stats.FrontHot, state.Stats.FrontWarm, state.Stats.FrontCold)
		case 1:
			return numberSection(gtx, th, "后区号码热度", state.Stats.BackHot, state.Stats.BackWarm, state.Stats.BackCold)
		case 2:
			return statsFooter(gtx, th)
		default:
			return layout.Dimensions{}
		}
	})
}

// numberSection 渲染一个分区，使用水平柱状图展示号码热度。
func numberSection(gtx layout.Context, th *theme.Theme, title string, hot, warm, cold []model.NumberFrequency) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Small, Bottom: th.Spacing.Small,
		Left: th.Spacing.Large, Right: th.Spacing.Large,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 卡片背景
		r := gtx.Dp(th.Shape.Medium)
		defer clip.RRect{
			Rect: image.Rect(0, 0, gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
			NE:   r, NW: r, SE: r, SW: r,
		}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Colors.Surface)

		return layout.UniformInset(th.Spacing.Medium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				// 分区标题
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(16), title)
					lbl.Font.Weight = font.Bold
					lbl.Color = th.Colors.OnSurface
					return lbl.Layout(gtx)
				}),
				// 柱状图：热号
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return statChartRow(gtx, th, "热号", hot, th.Colors.ChartOrange)
				}),
				// 柱状图：温号
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return statChartRow(gtx, th, "温号", warm, th.Colors.ChartGreen)
				}),
				// 柱状图：冷号
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return statChartRow(gtx, th, "冷号", cold, th.Colors.ChartBlue)
				}),
			)
		})
	})
}

// statChartRow 渲染一栏（含标签 + 水平柱状图）。
func statChartRow(gtx layout.Context, th *theme.Theme, label string, freqs []model.NumberFrequency, barColor color.NRGBA) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.XSmall,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 找到本组最大频次
		maxCount := 1
		for _, f := range freqs {
			if f.Count > maxCount {
				maxCount = f.Count
			}
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// 标签行
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th.Theme, unit.Sp(12), label)
				lbl.Font.Weight = font.Medium
				lbl.Color = th.Colors.OnSurface
				return lbl.Layout(gtx)
			}),
			// 柱状图（最多显示 8 个）
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				limit := 8
				if len(freqs) < limit {
					limit = len(freqs)
				}
				items := make([]lotwidget.BarItem, 0, limit)
				for _, f := range freqs[:limit] {
					items = append(items, lotwidget.BarItem{
						Label:    fmt.Sprintf("%02d", f.Number),
						Freq:     f.Count,
						MaxFreq:  maxCount,
						BarColor: barColor,
						FreqText: fmt.Sprintf("%d次", f.Count),
					})
				}
				return lotwidget.HorizontalBars(gtx, th, items)
			}),
		)
	})
}

// statsFooter 渲染统计页底部信息。
func statsFooter(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Medium, Bottom: th.Spacing.Large,
		Left: th.Spacing.Medium, Right: th.Spacing.Medium,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(th.Theme, unit.Sp(11), "基于大乐透历史开奖数据统计，仅供参考")
		lbl.Color = th.Colors.Disabled
		return lbl.Layout(gtx)
	})
}

// fetchStatsAsync 异步拉取并统计数据。
func fetchStatsAsync(state *StatsState, svc *Services, drawsCache *[]model.Draw) {
	defer func() {
		if svc.Invalidate != nil {
			svc.Invalidate()
		}
	}()

	// 先确保有开奖数据
	if len(*drawsCache) == 0 {
		draws, err := svc.Lottery.FetchDraws(context.Background(), 100)
		if err != nil {
			slog.Error("拉取统计数据失败", "error", err)
			state.Error = err
			state.Loading = false
			return
		}
		*drawsCache = draws
	}

	stats, err := svc.Stats.CalculateStats(context.Background(), *drawsCache, state.PeriodRange)
	if err != nil {
		slog.Error("冷热统计计算失败", "error", err)
		state.Error = err
		state.Loading = false
		return
	}

	state.Stats = stats
	state.Loaded = true
	state.Loading = false
	state.Error = nil
	slog.Info("冷热统计完成", "range", state.PeriodRange)
}
