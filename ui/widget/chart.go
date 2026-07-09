package widget

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/user/lottery/ui/theme"
)

// BarItem 表示单个水平柱状条的数据。
type BarItem struct {
	Label    string      // 号码，如 "01"
	Freq     int         // 出现次数
	MaxFreq  int         // 本组最大频次（用于计算条宽比例）
	BarColor color.NRGBA // 柱状条颜色
	FreqText string      // 频次文字，如 "8次"
}

// HorizontalBars 渲染一组水平柱状条。
func HorizontalBars(gtx layout.Context, th *theme.Theme, items []BarItem) layout.Dimensions {
	if len(items) == 0 {
		return layout.Dimensions{}
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, func() []layout.FlexChild {
		children := make([]layout.FlexChild, len(items))
		for i, item := range items {
			i, item := i, item
			children[i] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return barRow(gtx, th, item)
			})
		}
		return children
	}()...)
}

// barRow 渲染单行柱状条。
func barRow(gtx layout.Context, th *theme.Theme, item BarItem) layout.Dimensions {
	barH := gtx.Dp(unit.Dp(20))

	return layout.Inset{
		Top: unit.Dp(3), Bottom: unit.Dp(3),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// 号码标签（固定宽度 32dp）
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(34))
				gtx.Constraints.Max.X = gtx.Dp(unit.Dp(34))
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(13), item.Label)
					lbl.Font.Weight = font.Medium
					lbl.Color = th.Colors.OnSurface
					return lbl.Layout(gtx)
				})
			}),
			// 柱状条 + 频次
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					// 彩色柱状条
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						ratio := float32(item.Freq) / float32(item.MaxFreq)
						if ratio > 1 {
							ratio = 1
						}
						if ratio < 0.04 {
							ratio = 0.04
						}
						barW := int(float32(gtx.Constraints.Max.X) * ratio)
						if barW < barH {
							barW = barH
						}

						gtx.Constraints.Min.X = barW
						gtx.Constraints.Max.X = barW
						gtx.Constraints.Min.Y = barH
						gtx.Constraints.Max.Y = barH

						r := barH / 2
						defer clip.RRect{
							Rect: image.Rect(0, 0, barW, barH),
							NE:   r, NW: r, SE: r, SW: r,
						}.Push(gtx.Ops).Pop()
						paint.Fill(gtx.Ops, item.BarColor)
						return layout.Dimensions{Size: image.Pt(barW, barH)}
					}),
					// 频次文字
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{
							Left: unit.Dp(6),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(th.Theme, unit.Sp(12), item.FreqText)
							lbl.Color = th.Colors.Disabled
							return lbl.Layout(gtx)
						})
					}),
				)
			}),
		)
	})
}
