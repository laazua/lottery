// Package screen 提供各业务屏的布局编排逻辑。
package screen

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/user/lottery/model"
	"github.com/user/lottery/ui/theme"
	lotwidget "github.com/user/lottery/ui/widget"
)

// historyPageSize 每页显示的开奖期数。
const historyPageSize = 30

// HistoryState 维护历史开奖屏的状态（含所有 gioui 交互组件）。
type HistoryState struct {
	Draws   []model.Draw
	Total   int // 总记录数
	Page    int // 当前页码，从 1 开始
	Loading bool
	Loaded  bool // 是否已触发过首次加载
	Error   error

	// gioui 交互组件（必须持久化！）
	RefreshBtn widget.Clickable
	List       layout.List
	PrevBtn    widget.Clickable // 上一页
	NextBtn    widget.Clickable // 下一页
}

// HistoryLayout 渲染历史开奖查询页面。
func HistoryLayout(gtx layout.Context, th *theme.Theme, state *HistoryState, svc *Services, drawsCache *[]model.Draw) layout.Dimensions {
	// ═══ ① 事件检测（布局代码之前）═══
	// 修复 gioui layout.List 零值默认 Horizontal 导致列表不可见的问题。
	if state.List.Axis != layout.Vertical {
		state.List.Axis = layout.Vertical
	}

	if !state.Loaded && !state.Loading {
		state.Page = 1
		state.Loading = true
		state.Loaded = true
		state.Error = nil
		go fetchDrawsPageAsync(state, svc)
	}

	if state.RefreshBtn.Clicked(gtx) {
		state.Loading = true
		state.Error = nil
		go fetchDrawsPageAsync(state, svc)
	}

	// 上一页
	if state.PrevBtn.Clicked(gtx) && state.Page > 1 {
		state.Page--
		state.Loading = true
		state.Error = nil
		go fetchDrawsPageAsync(state, svc)
	}

	// 下一页
	if state.NextBtn.Clicked(gtx) && state.Page < totalPages(state) {
		state.Page++
		state.Loading = true
		state.Error = nil
		go fetchDrawsPageAsync(state, svc)
	}

	// ═══ ② 页面布局 ═══
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		// 顶部标题栏
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return historyHeader(gtx, th, state)
		}),
		// 表格内容（表头 + 数据行）
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return historyContent(gtx, th, state)
		}),
		// 底部分页栏
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return historyPagination(gtx, th, state)
		}),
	)
}

// formatAmount 将金额（元）格式化为人类可读的字符串。
func formatAmount(amount int64) string {
	switch {
	case amount >= 1_0000_0000: // >= 1亿
		v := float64(amount) / 1_0000_0000.0
		return fmt.Sprintf("%.1f亿", v)
	case amount >= 1_0000: // >= 1万
		v := float64(amount) / 1_0000.0
		return fmt.Sprintf("%.1f万", v)
	default:
		return fmt.Sprintf("%d元", amount)
	}
}

// totalPages 计算总页数。
func totalPages(state *HistoryState) int {
	if state.Total <= 0 {
		return 1
	}
	return (state.Total + historyPageSize - 1) / historyPageSize
}

// historyHeader 渲染顶部标题栏（含当期销售额、奖池总额和刷新按钮）。
func historyHeader(gtx layout.Context, th *theme.Theme, state *HistoryState) layout.Dimensions {
	return layout.Inset{
		Top: th.Spacing.Medium, Bottom: th.Spacing.Small,
		Left: th.Spacing.Large, Right: th.Spacing.Large,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			// 左侧标题
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th.Theme, unit.Sp(18), "开奖查询")
				lbl.Font.Weight = font.Bold
				lbl.Color = th.Colors.OnSurface
				return lbl.Layout(gtx)
			}),
			// 中间：当期销售额 + 奖池总额（红色字体）
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if len(state.Draws) == 0 {
					return layout.Dimensions{}
				}
				latest := state.Draws[0]
				saleStr := formatAmount(latest.SaleAmount)
				poolStr := formatAmount(latest.PoolAmount)
				info := fmt.Sprintf("销售额: %s  奖池: %s", saleStr, poolStr)
				lbl := material.Label(th.Theme, unit.Sp(11), info)
				lbl.Color = th.Colors.Error
				lbl.Font.Weight = font.Medium
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
			// 右侧刷新按钮
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				btn := lotwidget.FilledBtn(th, "刷新", &state.RefreshBtn)
				return btn.Layout(gtx)
			}),
		)
	})
}

// historyContent 渲染表格内容或空状态/加载状态。
// 列表第一项为表格表头，后续为数据行。
func historyContent(gtx layout.Context, th *theme.Theme, state *HistoryState) layout.Dimensions {
	if state.Loading && len(state.Draws) == 0 {
		return lotwidget.LoadingSkeleton(gtx, th)
	}

	if state.Error != nil && len(state.Draws) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(16), "数据加载失败")
					lbl.Color = th.Colors.Error
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th.Theme, unit.Sp(12), state.Error.Error())
					lbl.Color = th.Colors.Disabled
					return lbl.Layout(gtx)
				}),
			)
		})
	}

	if len(state.Draws) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th.Theme, unit.Sp(16), "暂无开奖数据")
			lbl.Color = th.Colors.Disabled
			return lbl.Layout(gtx)
		})
	}

	// 表格：总项数 = 1（表头）+ len(state.Draws)（数据行）
	totalItems := 1 + len(state.Draws)
	return state.List.Layout(gtx, totalItems, func(gtx layout.Context, index int) layout.Dimensions {
		if index == 0 {
			return drawTableHeader(gtx, th)
		}
		rowIdx := index - 1
		even := rowIdx%2 == 1 // 表头行不计入交替色
		return drawTableRow(gtx, th, state.Draws[rowIdx], even)
	})
}

// drawTableHeader 渲染表格表头行（灰色背景 + 四列标签）。
func drawTableHeader(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	return layout.Inset{
		Left:  th.Spacing.Small,
		Right: th.Spacing.Small,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 浅灰背景
		defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, th.Colors.TableHeaderBg)

		return layout.UniformInset(th.Spacing.Small).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(0.18, alignLeft(th, unit.Sp(11), "期号", font.Bold, th.Colors.Disabled)),
				layout.Flexed(0.26, alignLeft(th, unit.Sp(11), "日期", font.Bold, th.Colors.Disabled)),
				layout.Flexed(0.34, alignLeft(th, unit.Sp(11), "前区", font.Bold, th.Colors.Disabled)),
				layout.Flexed(0.22, alignLeft(th, unit.Sp(11), "后区", font.Bold, th.Colors.Disabled)),
			)
		})
	})
}

// drawTableRow 渲染单行开奖数据（交替色 + 底部分割线）。
func drawTableRow(gtx layout.Context, th *theme.Theme, draw model.Draw, alt bool) layout.Dimensions {
	return layout.Inset{
		Left:  th.Spacing.Small,
		Right: th.Spacing.Small,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 行背景（交替色）— 立即 Pop clip，避免影响后续渲染
		bg := th.Colors.Surface
		if alt {
			bg = th.Colors.TableAltBg
		}
		bgClip := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, bg)
		bgClip.Pop()

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			// 文本行
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(th.Spacing.Small).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
					}.Layout(gtx,
						layout.Flexed(0.18, alignLeft(th, unit.Sp(12), draw.Issue, font.Medium, th.Colors.OnSurface)),
						layout.Flexed(0.26, alignLeft(th, unit.Sp(12), draw.DrawTime.Format("2006-01-02"), font.Normal, th.Colors.OnSurface)),
						layout.Flexed(0.34, alignLeft(th, unit.Sp(12), formatFrontNums(draw), font.Medium, th.Colors.FrontBall)),
						layout.Flexed(0.22, alignLeft(th, unit.Sp(12), formatBackNums(draw), font.Medium, th.Colors.BackBall)),
					)
				})
			}),
			// 底部分割线（1px）— 独立 clip 只覆盖分割线区域
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lineH := gtx.Dp(unit.Dp(1))
				gtx.Constraints.Min.Y = lineH
				gtx.Constraints.Max.Y = lineH
				divClip := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
				paint.Fill(gtx.Ops, th.Colors.TableDivider)
				divClip.Pop()
				return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, lineH)}
			}),
		)
	})
}

// alignLeft 创建左对齐文本标签的辅助函数，返回 Widget 供 Flexed 使用。
func alignLeft(th *theme.Theme, size unit.Sp, txt string, weight font.Weight, txtColor color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(th.Theme, size, txt)
		lbl.Font.Weight = weight
		lbl.Color = txtColor
		return lbl.Layout(gtx)
	}
}

// formatFrontNums 将前区 5 个号码格式化为 "15 20 27 28 35"。
func formatFrontNums(draw model.Draw) string {
	parts := make([]string, 5)
	for i, n := range draw.FrontNumbers {
		parts[i] = fmt.Sprintf("%02d", n)
	}
	return strings.Join(parts, " ")
}

// formatBackNums 将后区 2 个号码格式化为 "02 11"。
func formatBackNums(draw model.Draw) string {
	parts := make([]string, 2)
	for i, n := range draw.BackNumbers {
		parts[i] = fmt.Sprintf("%02d", n)
	}
	return strings.Join(parts, " ")
}

// historyPagination 渲染底部分页栏。
func historyPagination(gtx layout.Context, th *theme.Theme, state *HistoryState) layout.Dimensions {
	page := state.Page
	total := totalPages(state)

	return layout.Inset{
		Top: th.Spacing.Small, Bottom: th.Spacing.Small,
		Left: th.Spacing.Large, Right: th.Spacing.Large,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// 分割线
		stack := clip.Rect(image.Rect(0, 0, gtx.Constraints.Max.X, 1)).Push(gtx.Ops)
		paint.Fill(gtx.Ops, th.Colors.TableDivider)
		stack.Pop()

		return layout.Inset{
			Top: th.Spacing.Small,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				// 上一页按钮
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(th.Theme, &state.PrevBtn, "◀ 上一页")
					btn.Background = th.Colors.Surface
					btn.Color = th.Colors.OnSurface
					if page <= 1 {
						btn.Background = th.Colors.DisabledBg
						btn.Color = th.Colors.Disabled
					}
					btn.CornerRadius = unit.Dp(14)
					btn.Inset = layout.Inset{
						Top: unit.Dp(4), Bottom: unit.Dp(4),
						Left: unit.Dp(10), Right: unit.Dp(10),
					}
					btn.TextSize = unit.Sp(13)
					return btn.Layout(gtx)
				}),
				// 页码信息
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					info := fmt.Sprintf("第 %d / %d 页", page, total)
					if state.Loading {
						info = "加载中..."
					}
					lbl := material.Label(th.Theme, unit.Sp(13), info)
					lbl.Color = th.Colors.OnSurface
					lbl.Font.Weight = font.Medium
					lbl.Alignment = text.Middle
					return layout.Center.Layout(gtx, lbl.Layout)
				}),
				// 下一页按钮
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := material.Button(th.Theme, &state.NextBtn, "下一页 ▶")
					btn.Background = th.Colors.Surface
					btn.Color = th.Colors.OnSurface
					if page >= total {
						btn.Background = th.Colors.DisabledBg
						btn.Color = th.Colors.Disabled
					}
					btn.CornerRadius = unit.Dp(14)
					btn.Inset = layout.Inset{
						Top: unit.Dp(4), Bottom: unit.Dp(4),
						Left: unit.Dp(10), Right: unit.Dp(10),
					}
					btn.TextSize = unit.Sp(13)
					return btn.Layout(gtx)
				}),
			)
		})
	})
}

// fetchDrawsPageAsync 在 goroutine 中异步拉取分页开奖数据。
func fetchDrawsPageAsync(state *HistoryState, svc *Services) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("拉取开奖数据 panic", "page", state.Page, "panic", r)
			state.Error = fmt.Errorf("内部错误: %v", r)
		}
		state.Loading = false
		if svc.Invalidate != nil {
			svc.Invalidate()
		}
	}()

	page, err := svc.Lottery.FetchDrawsPage(context.Background(), state.Page, historyPageSize)
	if err != nil {
		slog.Error("拉取开奖数据失败", "page", state.Page, "error", err)
		state.Error = err
		return
	}

	slog.Info("拉取开奖数据成功", "count", len(page.Draws), "total", page.Total, "page", state.Page)
	state.Draws = page.Draws
	state.Total = page.Total
	state.Error = nil
}
