package widget

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/user/lottery/ui/theme"
)

// LoadingSkeleton 显示加载中的骨架屏占位。
func LoadingSkeleton(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(th.Theme, unit.Sp(16), "加载中...")
		label.Color = th.Colors.Disabled
		return label.Layout(gtx)
	})
}

// LoadingIndicator 在页面顶部显示一个简短的加载提示。
func LoadingIndicator(gtx layout.Context, th *theme.Theme, msg string) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Medium, Bottom: th.Spacing.Medium,
		Left: th.Spacing.Medium, Right: th.Spacing.Medium,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(th.Theme, unit.Sp(14), msg)
		lbl.Color = th.Colors.Disabled
		return lbl.Layout(gtx)
	})
}
