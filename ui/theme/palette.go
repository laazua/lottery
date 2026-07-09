// Package theme 提供统一的视觉设计主题。
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

// Colors 定义基于参考 UI（ui.png）的调色板。
// 色值来源于 docs/10-ui-redesign-v2.md §2.1。
type Colors struct {
	Primary         color.NRGBA // #176EFB - Header、选中状态
	OnPrimary       color.NRGBA // #FFFFFF
	FrontBall       color.NRGBA // #F46E6A - 前区号码（红色）
	BackBall        color.NRGBA // #2665EF - 后区号码（蓝色）
	Surface         color.NRGBA // #FFFFFF - 卡片底色
	OnSurface       color.NRGBA // #1E1E20 - 主文字
	Bg              color.NRGBA // #E5DED2 - 页面背景（暖调米灰）
	Hot             color.NRGBA // #F46E6A (红)
	Warm            color.NRGBA // #FD9E37 (橙)
	Cold            color.NRGBA // #2E70F9 (蓝)
	Miss            color.NRGBA // #843CF8 (紫)
	ChartOrange     color.NRGBA // #FD9E37
	ChartGreen      color.NRGBA // #37C75F
	ChartPurple     color.NRGBA // #843CF8
	ChartBlue       color.NRGBA // #2E70F9
	NavSelectedBg   color.NRGBA // #D8E5FD - 选中 Tab 背景
	NavUnselectedBg color.NRGBA // #EFF4FD - 未选中 Tab 背景
	Error           color.NRGBA // #F46E6A
	Disabled        color.NRGBA // #8C8C8C
	DisabledBg      color.NRGBA // #E0E0E0
	Outline         color.NRGBA // #D8E5FD - 卡片描边/分割线
	TableHeaderBg   color.NRGBA // #F5F5F5 - 表格表头背景
	TableAltBg      color.NRGBA // #F0F4FA - 表格交替行背景
	TableDivider    color.NRGBA // #E8ECF0 - 表格行分割线
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
	Ball   unit.Dp
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

// NewTheme 创建基于参考 UI 的新主题。
func NewTheme() *Theme {
	th := material.NewTheme()
	th.TextSize = unit.Sp(14)

	return &Theme{
		Theme: th,
		Colors: Colors{
			Primary:         color.NRGBA{R: 0x17, G: 0x6E, B: 0xFB, A: 0xFF}, // #176EFB
			OnPrimary:       color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
			FrontBall:       color.NRGBA{R: 0xF4, G: 0x6E, B: 0x6A, A: 0xFF}, // #F46E6A
			BackBall:        color.NRGBA{R: 0x26, G: 0x65, B: 0xEF, A: 0xFF}, // #2665EF
			Surface:         color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, // #FFFFFF
			OnSurface:       color.NRGBA{R: 0x1E, G: 0x1E, B: 0x20, A: 0xFF}, // #1E1E20
			Bg:              color.NRGBA{R: 0xE5, G: 0xDE, B: 0xD2, A: 0xFF}, // #E5DED2
			Hot:             color.NRGBA{R: 0xF4, G: 0x6E, B: 0x6A, A: 0xFF}, // #F46E6A
			Warm:            color.NRGBA{R: 0xFD, G: 0x9E, B: 0x37, A: 0xFF}, // #FD9E37
			Cold:            color.NRGBA{R: 0x2E, G: 0x70, B: 0xF9, A: 0xFF}, // #2E70F9
			Miss:            color.NRGBA{R: 0x84, G: 0x3C, B: 0xF8, A: 0xFF}, // #843CF8
			ChartOrange:     color.NRGBA{R: 0xFD, G: 0x9E, B: 0x37, A: 0xFF},
			ChartGreen:      color.NRGBA{R: 0x37, G: 0xC7, B: 0x5F, A: 0xFF},
			ChartPurple:     color.NRGBA{R: 0x84, G: 0x3C, B: 0xF8, A: 0xFF},
			ChartBlue:       color.NRGBA{R: 0x2E, G: 0x70, B: 0xF9, A: 0xFF},
			NavSelectedBg:   color.NRGBA{R: 0xD8, G: 0xE5, B: 0xFD, A: 0xFF}, // #D8E5FD
			NavUnselectedBg: color.NRGBA{R: 0xEF, G: 0xF4, B: 0xFD, A: 0xFF}, // #EFF4FD
			Error:           color.NRGBA{R: 0xF4, G: 0x6E, B: 0x6A, A: 0xFF},
			Disabled:        color.NRGBA{R: 0x8C, G: 0x8C, B: 0x8C, A: 0xFF}, // #8C8C8C
			DisabledBg:      color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF},
			Outline:         color.NRGBA{R: 0xD8, G: 0xE5, B: 0xFD, A: 0xFF}, // #D8E5FD
			TableHeaderBg:   color.NRGBA{R: 0xF5, G: 0xF5, B: 0xF5, A: 0xFF}, // #F5F5F5
			TableAltBg:      color.NRGBA{R: 0xF0, G: 0xF4, B: 0xFA, A: 0xFF}, // #F0F4FA
			TableDivider:    color.NRGBA{R: 0xE8, G: 0xEC, B: 0xF0, A: 0xFF}, // #E8ECF0
		},
		Spacing: Spacing{
			XXSmall: unit.Dp(2),
			XSmall:  unit.Dp(4),
			Small:   unit.Dp(8),
			Medium:  unit.Dp(12),
			Large:   unit.Dp(16),
			XLarge:  unit.Dp(24),
		},
		Shape: Shape{
			Small:  unit.Dp(6),
			Medium: unit.Dp(10),
			Large:  unit.Dp(14),
			Ball:   unit.Dp(20),
		},
		Elevation: Elevation{
			Flat:   0,
			Low:    unit.Dp(1),
			Medium: unit.Dp(2),
			High:   unit.Dp(4),
		},
		BallSizes: BallSizes{
			Small:  unit.Dp(15),
			Medium: unit.Dp(22),
			Large:  unit.Dp(28),
		},
	}
}
