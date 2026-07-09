package screen

import (
	"image"
	"math/rand"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/user/lottery/ui/theme"
	lotwidget "github.com/user/lottery/ui/widget"
)

// RandomState 维护随机选号屏的状态。
type RandomState struct {
	FrontNums [5]int
	BackNums  [2]int
	Generated bool

	GenerateBtn widget.Clickable
}

// RandomLayout 渲染随机选号页面。
func RandomLayout(gtx layout.Context, th *theme.Theme, state *RandomState) layout.Dimensions {
	if state.GenerateBtn.Clicked(gtx) {
		state.FrontNums = generateRandomFront()
		state.BackNums = generateRandomBack()
		state.Generated = true
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return randomHeader(gtx, th, state)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return randomContent(gtx, th, state)
		}),
	)
}

// randomHeader 渲染标题和生成按钮。
func randomHeader(gtx layout.Context, th *theme.Theme, state *RandomState) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Medium, Bottom: th.Spacing.Small,
		Left: th.Spacing.Large, Right: th.Spacing.Large,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Horizontal,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th.Theme, unit.Sp(18), "随机选号")
				lbl.Font.Weight = font.Bold
				lbl.Color = th.Colors.OnSurface
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := lotwidget.FilledBtn(th, "生成随机", &state.GenerateBtn)
				return btn.Layout(gtx)
			}),
		)
	})
}

// randomContent 渲染号码展示区域。
func randomContent(gtx layout.Context, th *theme.Theme, state *RandomState) layout.Dimensions {
	if !state.Generated {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.Theme, unit.Sp(16), "点击上方按钮生成随机号码")
			lbl.Color = th.Colors.Disabled
			return lbl.Layout(gtx)
		})
	}

	return layout.Inset{
		Top:  th.Spacing.Large,
		Left: th.Spacing.Large, Right: th.Spacing.Large,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// 前区卡片
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return randomBallCard(gtx, th, "前区号码", state.FrontNums[:], lotwidget.BallFront)
			}),
			// 后区卡片
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return randomBallCard(gtx, th, "后区号码", state.BackNums[:], lotwidget.BallBack)
			}),
			// 提示
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top: th.Spacing.Large,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(11), "仅供娱乐，理性购彩")
					lbl.Color = th.Colors.Disabled
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

// randomBallCard 渲染号码卡片（含标题和号码球行）。
func randomBallCard(gtx layout.Context, th *theme.Theme, title string, nums []int, status lotwidget.BallStatus) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Small, Bottom: th.Spacing.Small,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(15), title)
					lbl.Font.Weight = font.Bold
					lbl.Color = th.Colors.OnSurface
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{
						Top: th.Spacing.Small,
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{
							Axis: layout.Horizontal,
						}.Layout(gtx, randomBallRow(th, nums, status)...)
					})
				}),
			)
		})
	})
}

// randomBallRow 渲染号码球行。
func randomBallRow(th *theme.Theme, nums []int, status lotwidget.BallStatus) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(nums))
	for _, n := range nums {
		n := n
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Right: th.Spacing.Small,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return lotwidget.Ball(gtx, th, n, status, th.BallSizes.Large)
			})
		}))
	}
	return children
}

// generateRandomFront 生成 5 个不重复的前区随机号码（1-35）。
func generateRandomFront() [5]int {
	nums := rand.Perm(35)[:5]
	for i := range nums {
		nums[i]++
	}
	// 冒泡排序
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			if nums[i] > nums[j] {
				nums[i], nums[j] = nums[j], nums[i]
			}
		}
	}
	var result [5]int
	copy(result[:], nums)
	return result
}

// generateRandomBack 生成 2 个不重复的后区随机号码（1-12）。
func generateRandomBack() [2]int {
	nums := rand.Perm(12)[:2]
	for i := range nums {
		nums[i]++
	}
	if nums[0] > nums[1] {
		nums[0], nums[1] = nums[1], nums[0]
	}
	var result [2]int
	copy(result[:], nums)
	return result
}
