package screen

import (
	"context"
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

// RecommendState 维护推荐选号屏的状态。
type RecommendState struct {
	Recommendation *model.Recommendation
	Loading        bool
	Loaded         bool
	Error          error

	// gioui 交互组件（持久化）
	GenerateBtn widget.Clickable
}

// RecommendLayout 渲染智能推荐页面。
func RecommendLayout(gtx layout.Context, th *theme.Theme, state *RecommendState, svc *Services, drawsCache *[]model.Draw) layout.Dimensions {
	// ═══ ① 事件检测 ═══
	if state.GenerateBtn.Clicked(gtx) {
		state.Loading = true
		go generateRecommendAsync(state, svc, drawsCache)
	}

	// ═══ ② 页面布局 ═══
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return recommendHeader(gtx, th, state, svc)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return recommendContent(gtx, th, state)
		}),
	)
}

// recommendHeader 渲染推荐页标题和生成按钮。
func recommendHeader(gtx layout.Context, th *theme.Theme, state *RecommendState, svc *Services) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Medium, Bottom: th.Spacing.Small,
		Left: th.Spacing.Large, Right: th.Spacing.Large,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Horizontal,
		}.Layout(gtx,
			// 标题
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th.Theme, unit.Sp(18), "智能推荐")
				lbl.Font.Weight = font.Bold
				lbl.Color = th.Colors.OnSurface
				return lbl.Layout(gtx)
			}),
			// 生成按钮
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := lotwidget.FilledBtn(th, "生成推荐", &state.GenerateBtn)
				return btn.Layout(gtx)
			}),
		)
	})
}

// recommendContent 渲染推荐结果或占位提示。
func recommendContent(gtx layout.Context, th *theme.Theme, state *RecommendState) layout.Dimensions {
	if state.Loading {
		return lotwidget.LoadingSkeleton(gtx, th)
	}

	if state.Error != nil {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(16), "推荐生成失败")
					lbl.Color = th.Colors.Error
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(12), "请检查数据源后重试")
					lbl.Color = th.Colors.Disabled
					return lbl.Layout(gtx)
				}),
			)
		})
	}

	if state.Recommendation == nil {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.Theme, unit.Sp(16), "点击上方按钮生成推荐号码")
			lbl.Color = th.Colors.Disabled
			return lbl.Layout(gtx)
		})
	}

	// ═══ 渲染推荐结果 ═══
	return layout.Inset{
		Top:   th.Spacing.Small,
		Left:  th.Spacing.Large,
		Right: th.Spacing.Large,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// 前区
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return recommendCard(gtx, th, "前区推荐", state.Recommendation.FrontNumbers)
			}),
			// 后区
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return recommendCard(gtx, th, "后区推荐", state.Recommendation.BackNumbers)
			}),
			// 提示
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top: th.Spacing.Large,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(11), "仅供参考，理性购彩")
					lbl.Color = th.Colors.Disabled
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

// recommendCard 渲染推荐号码卡片。
func recommendCard(gtx layout.Context, th *theme.Theme, title string, nums []model.RecommendNumber) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Small, Bottom: th.Spacing.Small,
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
				// 标题
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(15), title)
					lbl.Font.Weight = font.Bold
					lbl.Color = th.Colors.OnSurface
					return lbl.Layout(gtx)
				}),
				// 号码行
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top: th.Spacing.Small,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis: layout.Horizontal,
						}.Layout(gtx, recommendBallRow(th, nums)...)
					})
				}),
			)
		})
	})
}

// recommendBallRow 渲染推荐号码球行（含推荐理由标签）。
func recommendBallRow(th *theme.Theme, nums []model.RecommendNumber) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(nums))
	for _, n := range nums {
		n := n
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Right: th.Spacing.Small,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					// 号码球
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return lotwidget.Ball(gtx, th, n.Number, recommendStatus(n.Reason), th.BallSizes.Large)
					}),
					// 理由标签
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(th.Theme, unit.Sp(10), n.Reason)
						lbl.Color = th.Colors.Disabled
						return lbl.Layout(gtx)
					}),
				)
			})
		}))
	}
	return children
}

// recommendStatus 根据推荐理由标签返回 BallStatus。
// 前区号码使用 BallFront（红色），后区号码使用 BallBack（蓝色）。
// 由于 recommend.go 不区分前后区，这里根据理由映射。
func recommendStatus(reason string) lotwidget.BallStatus {
	switch reason {
	case "热号":
		return lotwidget.BallHot
	case "温号":
		return lotwidget.BallWarm
	case "遗漏":
		return lotwidget.BallMiss
	default:
		return lotwidget.BallFront
	}
}

// generateRecommendAsync 异步生成推荐号码。
func generateRecommendAsync(state *RecommendState, svc *Services, drawsCache *[]model.Draw) {
	defer func() {
		state.Loading = false
		if svc.Invalidate != nil {
			svc.Invalidate()
		}
	}()

	// 先确保有开奖数据用作统计
	if len(*drawsCache) == 0 {
		draws, err := svc.Lottery.FetchDraws(context.Background(), 100)
		if err != nil {
			slog.Error("拉取推荐数据失败", "error", err)
			state.Error = err
			return
		}
		*drawsCache = draws
	}

	stats, err := svc.Stats.CalculateStats(context.Background(), *drawsCache, 50)
	if err != nil {
		slog.Error("推荐统计计算失败", "error", err)
		state.Error = err
		return
	}

	rec, err := svc.Recommend.GenerateRecommendation(context.Background(), stats)
	if err != nil {
		slog.Error("推荐生成失败", "error", err)
		state.Error = err
		return
	}

	state.Recommendation = rec
	state.Error = nil
	slog.Info("推荐生成完成", "frontCount", len(rec.FrontNumbers), "backCount", len(rec.BackNumbers))
}
