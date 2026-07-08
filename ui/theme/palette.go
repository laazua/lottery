// Package theme 提供统一的视觉设计主题，基于 Material Design 3 配色体系。
package theme

import (
	"image/color"

	"gioui.org/unit"
	"gioui.org/widget/material"
)

// Theme 扩展 material.Theme，包含完整的设计 Token。
type Theme struct {
	*material.Theme
	Colors    Colors
	Spacing   Spacing
	Shape     Shape
	Elevation Elevation
	BallSizes BallSizes
}

// Colors 定义 Material Design 3 调色板。
type Colors struct {
	// 主色
	Primary   color.NRGBA
	OnPrimary color.NRGBA
	Secondary color.NRGBA
	Error     color.NRGBA

	// 表面
	Surface   color.NRGBA // 卡片底色
	OnSurface color.NRGBA // 卡片上文字
	Bg        color.NRGBA // 页面背景

	// 冷热主题
	Hot  color.NRGBA
	Warm color.NRGBA
	Cold color.NRGBA
	Miss color.NRGBA

	// 功能色
	Disabled   color.NRGBA
	DisabledBg color.NRGBA
	Outline    color.NRGBA // 分割线
}

// Spacing 定义间距层级。
type Spacing struct {
	XXSmall unit.Dp
	XSmall  unit.Dp
	Small   unit.Dp
	Medium  unit.Dp
	Large   unit.Dp
	XLarge  unit.Dp
}

// Shape 定义圆角尺寸。
type Shape struct {
	Small  unit.Dp
	Medium unit.Dp
	Large  unit.Dp
	Ball   unit.Dp // 号码球圆形
}

// Elevation 定义阴影层级。
type Elevation struct {
	Flat   unit.Dp
	Low    unit.Dp
	Medium unit.Dp
	High   unit.Dp
}

// BallSizes 定义号码球尺寸。
type BallSizes struct {
	Small  unit.Dp
	Medium unit.Dp
	Large  unit.Dp
}

// NewTheme 创建默认主题（Material Design 3 风格）。
func NewTheme() *Theme {
	th := material.NewTheme()
	th.TextSize = unit.Sp(16)

	return &Theme{
		Theme: th,
		Colors: Colors{
			Primary:   color.NRGBA{R: 0xE8, G: 0x3E, B: 0x3C, A: 0xFF}, // Red600
			OnPrimary: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
			Secondary: color.NRGBA{R: 0x1E, G: 0x88, B: 0xE5, A: 0xFF}, // Blue600
			Error:     color.NRGBA{R: 0xB7, G: 0x1C, B: 0x1C, A: 0xFF},
			Surface:   color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
			OnSurface: color.NRGBA{R: 0x1C, G: 0x1B, B: 0x1F, A: 0xFF},
			Bg:        color.NRGBA{R: 0xF5, G: 0xF5, B: 0xF5, A: 0xFF},
			Hot:       color.NRGBA{R: 0xE5, G: 0x39, B: 0x35, A: 0xFF},
			Warm:      color.NRGBA{R: 0xFB, G: 0x8C, B: 0x00, A: 0xFF},
			Cold:      color.NRGBA{R: 0x21, G: 0x96, B: 0xF3, A: 0xFF},
			Miss:      color.NRGBA{R: 0x9C, G: 0x27, B: 0xB0, A: 0xFF},
			Disabled:  color.NRGBA{R: 0x9E, G: 0x9E, B: 0x9E, A: 0xFF},
			DisabledBg: color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF},
			Outline:   color.NRGBA{R: 0xDC, G: 0xDC, B: 0xDC, A: 0xFF},
		},
		Spacing: Spacing{
			XXSmall: unit.Dp(2),
			XSmall:  unit.Dp(4),
			Small:   unit.Dp(6),
			Medium:  unit.Dp(12),
			Large:   unit.Dp(16),
			XLarge:  unit.Dp(24),
		},
		Shape: Shape{
			Small:  unit.Dp(8),
			Medium: unit.Dp(12),
			Large:  unit.Dp(16),
			Ball:   unit.Dp(20),
		},
		Elevation: Elevation{
			Flat:   0,
			Low:    unit.Dp(1),
			Medium: unit.Dp(2),
			High:   unit.Dp(4),
		},
		BallSizes: BallSizes{
			Small:  unit.Dp(28),
			Medium: unit.Dp(36),
			Large:  unit.Dp(44),
		},
	}
}
