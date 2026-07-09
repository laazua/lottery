package widget

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/user/lottery/ui/theme"
)

// FilledBtn 渲染主色填充的胶囊形按钮。
func FilledBtn(th *theme.Theme, label string, btn *widget.Clickable) material.ButtonStyle {
	style := material.Button(th.Theme, btn, label)
	style.Background = th.Colors.Primary
	style.Color = th.Colors.OnPrimary
	style.CornerRadius = unit.Dp(16)
	style.Inset = layout.Inset{
		Top: unit.Dp(6), Bottom: unit.Dp(6),
		Left: unit.Dp(14), Right: unit.Dp(14),
	}
	style.TextSize = unit.Sp(13)
	return style
}

// SmallFilledBtn 渲染紧凑型填充按钮。
// 适用于"近20期""近50期"等选项按钮。
func SmallFilledBtn(th *theme.Theme, label string, btn *widget.Clickable) material.ButtonStyle {
	style := material.Button(th.Theme, btn, label)
	style.Background = th.Colors.Primary
	style.Color = th.Colors.OnPrimary
	style.CornerRadius = unit.Dp(14)
	style.Inset = layout.Inset{
		Top: unit.Dp(4), Bottom: unit.Dp(4),
		Left: unit.Dp(10), Right: unit.Dp(10),
	}
	style.TextSize = unit.Sp(13)
	return style
}

// OutlineBtn 渲染描边胶囊形按钮。
// 适用于"近20期""近50期"等非选中态选项。
func OutlineBtn(th *theme.Theme, label string, btn *widget.Clickable) material.ButtonStyle {
	style := material.Button(th.Theme, btn, label)
	style.Background = th.Colors.Surface
	style.Color = th.Colors.OnSurface
	style.CornerRadius = unit.Dp(14)
	style.Inset = layout.Inset{
		Top: unit.Dp(4), Bottom: unit.Dp(4),
		Left: unit.Dp(10), Right: unit.Dp(10),
	}
	style.TextSize = unit.Sp(13)
	return style
}
