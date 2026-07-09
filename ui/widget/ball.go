// Package widget 提供大乐透 APP 可复用的 UI 组件。
package widget

import (
	"image"
	"image/color"
	"strconv"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/user/lottery/ui/theme"
)

// BallStatus 定义号码球的显示状态。
type BallStatus int

const (
	BallNormal BallStatus = iota
	BallFront             // 前区号码（红色 #F46E6A）
	BallBack              // 后区号码（蓝色 #2665EF）
	BallHot
	BallWarm
	BallCold
	BallMiss
)

// Ball 渲染单个号码球。
// status 控制颜色，size 控制球径。
func Ball(gtx layout.Context, th *theme.Theme, number int, status BallStatus, size unit.Dp) layout.Dimensions {
	diameter := gtx.Dp(size)

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints = layout.Exact(image.Pt(diameter, diameter))

		// 球体背景
		rect := image.Rect(0, 0, diameter, diameter)
		rOps := clip.UniformRRect(rect, diameter/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, ballColor(th, status))
		rOps.Pop()

		// 高光（顶部 1/3 处半透明白）
		highlight := image.Rect(0, 0, diameter, diameter/2)
		hOps := clip.UniformRRect(highlight, diameter/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x24})
		hOps.Pop()

		// 号码文字
		label := material.Label(th.Theme, unit.Sp(size/2.5), strconv.Itoa(number))
		label.Color = th.Colors.OnPrimary
		label.Font.Weight = font.Bold
		label.Alignment = text.Middle
		return label.Layout(gtx)
	})
}

// ballColor 根据状态返回对应的颜色。
func ballColor(th *theme.Theme, status BallStatus) color.NRGBA {
	switch status {
	case BallFront:
		return th.Colors.FrontBall
	case BallBack:
		return th.Colors.BackBall
	case BallHot:
		return th.Colors.Hot
	case BallWarm:
		return th.Colors.Warm
	case BallCold:
		return th.Colors.Cold
	case BallMiss:
		return th.Colors.Miss
	default:
		return th.Colors.Primary
	}
}
