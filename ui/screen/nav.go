// Package screen 提供底部导航栏组件（在此包中避免与 widget 循环引用）。
package screen

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/user/lottery/ui/theme"
)

// TabInfo 定义单个 Tab 的显示信息。
type TabInfo struct {
	Icon  string
	Label string
}

// TabTitles 返回底部 Tab 栏三屏的配置。
func TabTitles() []TabInfo {
	return []TabInfo{
		{Icon: "📋", Label: "开奖查询"},
		{Icon: "📊", Label: "冷热统计"},
		{Icon: "🎯", Label: "智能推荐"},
	}
}

// BottomNavLayout 渲染 MD3 风格底部导航栏。
func BottomNavLayout(gtx layout.Context, th *theme.Theme, current ScreenID, btns *[3]widget.Clickable) layout.Dimensions {
	navH := gtx.Dp(unit.Dp(64))
	gtx.Constraints.Min.Y = navH
	gtx.Constraints.Max.Y = navH

	// 导航栏背景（带顶部圆角）
	defer clip.RRect{
		Rect: image.Rect(0, 0, gtx.Constraints.Max.X, navH),
		SE: gtx.Dp(th.Shape.Medium), SW: gtx.Dp(th.Shape.Medium),
	}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, th.Colors.Surface)

	// 顶部细线分割
	stack := clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, 1)).Push(gtx.Ops)
	paint.Fill(gtx.Ops, th.Colors.Outline)
	stack.Pop()

	titles := TabTitles()
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return tabItem(gtx, th, titles[0].Icon, titles[0].Label, current == ScreenHistory, &btns[0])
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return tabItem(gtx, th, titles[1].Icon, titles[1].Label, current == ScreenStats, &btns[1])
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return tabItem(gtx, th, titles[2].Icon, titles[2].Label, current == ScreenRecommend, &btns[2])
		}),
	)
}

// tabItem 渲染单个 Tab 项。
func tabItem(gtx layout.Context, th *theme.Theme, icon, label string, selected bool, btn *widget.Clickable) layout.Dimensions {
	return material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top: th.Spacing.XSmall, Bottom: th.Spacing.XXSmall,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				// 图标行
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(20), icon)
					if selected {
						lbl.Color = th.Colors.Primary
					} else {
						lbl.Color = th.Colors.Disabled
					}
					return lbl.Layout(gtx)
				}),
				// 标签行
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(10), label)
					lbl.Font.Weight = font.Medium
					if selected {
						lbl.Color = th.Colors.Primary
					} else {
						lbl.Color = th.Colors.Disabled
					}
					return lbl.Layout(gtx)
				}),
				// 选中指示条
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !selected {
						barH := gtx.Dp(unit.Dp(3))
						return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, barH)}
					}
					barW := gtx.Dp(unit.Dp(24))
					barH := gtx.Dp(unit.Dp(3))
					gtx.Constraints = layout.Exact(image.Pt(barW, barH))
					r := barH / 2
					defer clip.RRect{
						Rect: image.Rect(0, 0, barW, barH),
						NE: r, NW: r, SE: r, SW: r,
					}.Push(gtx.Ops).Pop()
					paint.Fill(gtx.Ops, th.Colors.Primary)
					return layout.Dimensions{Size: image.Pt(barW, barH)}
				}),
			)
		})
	})
}
