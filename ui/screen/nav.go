// Package screen 提供底部导航栏组件。
package screen

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
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

// TabTitles 返回底部 Tab 栏四屏的配置。
func TabTitles() []TabInfo {
	return []TabInfo{
		{Icon: "开", Label: "开奖查询"},
		{Icon: "冷", Label: "冷热统计"},
		{Icon: "推", Label: "智能推荐"},
		{Icon: "随", Label: "随机选号"},
	}
}

// BottomNavLayout 渲染带圆角背景的底部导航栏。
func BottomNavLayout(gtx layout.Context, th *theme.Theme, current ScreenID, btns *[4]widget.Clickable) layout.Dimensions {
	navH := gtx.Dp(unit.Dp(41))
	gtx.Constraints.Min.Y = navH
	gtx.Constraints.Max.Y = navH

	// 导航栏背景（白色），clip 约束到导航栏区域
	defer clip.RRect{
		Rect: image.Rect(0, 0, gtx.Constraints.Max.X, navH),
		NE:   gtx.Dp(th.Shape.Medium), NW: gtx.Dp(th.Shape.Medium),
		SE: 0, SW: 0,
	}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, th.Colors.Surface)

	// 顶部细线分割
	stack := clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, 1)).Push(gtx.Ops)
	paint.Fill(gtx.Ops, th.Colors.Outline)
	stack.Pop()

	titles := TabTitles()
	screenIDs := []ScreenID{ScreenHistory, ScreenStats, ScreenRecommend, ScreenRandom}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx, func() []layout.FlexChild {
		children := make([]layout.FlexChild, len(titles))
		for i := range titles {
			i := i
			children[i] = layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return tabItem(gtx, th, titles[i].Icon, titles[i].Label, current == screenIDs[i], &btns[i])
			})
		}
		return children
	}()...)
}

// tabItem 渲染单个 Tab 项（带选中圆角背景和圆形图标）。
func tabItem(gtx layout.Context, th *theme.Theme, icon, label string, selected bool, btn *widget.Clickable) layout.Dimensions {
	return material.Clickable(gtx, btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Left: unit.Dp(6), Right: unit.Dp(6),
			Top: unit.Dp(6), Bottom: unit.Dp(6),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// 选中时绘制圆角背景
			if selected {
				bgW := gtx.Constraints.Max.X
				bgH := gtx.Constraints.Max.Y
				r := bgH / 2
				if r > bgW/2 {
					r = bgW / 2
				}
				defer clip.RRect{
					Rect: image.Rect(0, 0, bgW, bgH),
					NE:   r, NW: r, SE: r, SW: r,
				}.Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, th.Colors.NavSelectedBg)
			}

			// Tab 内容：图标 + 标签
			return layout.Flex{
				Axis:      layout.Vertical,
				Alignment: layout.Middle,
			}.Layout(gtx,
				// 图标行（圆形背景中的中文字符）
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					iconSize := gtx.Dp(unit.Dp(24))
					return layout.Inset{
						Top: unit.Dp(2),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
						// 圆形背景
						defer clip.UniformRRect(image.Rect(0, 0, iconSize, iconSize), iconSize/2).Push(gtx.Ops).Pop()
						if selected {
							paint.Fill(gtx.Ops, th.Colors.Primary)
						} else {
							paint.Fill(gtx.Ops, th.Colors.DisabledBg)
						}
						// 中文字符
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(th.Theme, unit.Sp(12), icon)
							lbl.Color = th.Colors.OnPrimary
							lbl.Font.Weight = font.Bold
							lbl.Alignment = text.Middle
							return lbl.Layout(gtx)
						})
					})
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
			)
		})
	})
}
