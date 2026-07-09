# 12-开奖查询页面优化

## 优化背景

对照 `docs/11-history-page-layout.md` 布局规范和 `docs/01-architecture.md` 架构文档，对开奖查询页面进行了系统性的问题诊断、代码优化和缺陷修复。

## 1. 新增主题色 Token（palette.go）

### 问题

表头背景色 (`#F5F5F5`)、交替行色 (`#F0F4FA`)、分割线色 (`#E8ECF0`) 等表格专用色值均以硬编码 `color.NRGBA{...}` 散落在 `history.go` 各处，与项目主题化设计体系脱节。

### 方案

在 `ui/theme/palette.go` 的 `Colors` 结构体中新增三个专用 Token：

| Token | 色值 | 用途 | 文档依据 |
|---|---|---|---|
| `TableHeaderBg` | `#F5F5F5` | 表格表头背景 | 11-history §行样式.表头行 |
| `TableAltBg` | `#F0F4FA` | 表格交替行背景 | 11-history §行样式.数据行 |
| `TableDivider` | `#E8ECF0` | 表格行分割线 | 11-history §行样式.数据行 |

### 影响范围

- [ui/theme/palette.go:49-52](ui/theme/palette.go#L49-L52) — 结构体字段定义
- [ui/theme/palette.go:111-114](ui/theme/palette.go#L111-L114) — `NewTheme()` 初始化值

---

## 2. 消除硬编码颜色（history.go）

### 变更清单

| 位置 | 修改前 | 修改后 |
|---|---|---|
| 表头背景 | `color.NRGBA{R:0xF5,G:0xF5,B:0xF5,A:0xFF}` | `th.Colors.TableHeaderBg` |
| 表头文字 | `color.NRGBA{R:0x8C,G:0x8C,B:0x8C,A:0xFF}` | `th.Colors.Disabled` |
| 交替行背景 | `color.NRGBA{R:0xF0,G:0xF4,B:0xFA,A:0xFF}` | `th.Colors.TableAltBg` |
| 行分割线 | `color.NRGBA{R:0xE8,G:0xEC,B:0xF0,A:0xFF}` | `th.Colors.TableDivider` |
| 分页分割线 | `th.Colors.Outline` (#D8E5FD 蓝色调) | `th.Colors.TableDivider` |

其中分页分割线此前使用 `Outline` 色值（`#D8E5FD`，蓝色调），与文档规范 `#E8ECF0` 不符，一并修正。

---

## 3. 关键 Bug 修复

### 3.1 layout.List.Axis 零值陷阱（P0）

**影响屏**：开奖查询、冷热统计

**根因**：gioui v0.10.0 中 `layout.Axis` 定义为：

```go
const (
    Horizontal Axis = iota  // = 0
    Vertical                // = 1
)
```

`layout.List` 零值初始化时 `Axis` 默认为 `Horizontal`（值 0），导致列表**水平滚动**。表格数据行被推至屏幕右侧不可见区域，表现为"没有数据"。

**修复**：在 `HistoryLayout` 和 `StatsLayout` 顶部，一次性修正：

```go
if state.List.Axis != layout.Vertical {
    state.List.Axis = layout.Vertical
}
```

**影响文件**：
- [ui/screen/history.go:48-50](ui/screen/history.go#L48-L50)
- [ui/screen/stats.go:52-54](ui/screen/stats.go#L52-L54)

### 3.2 defer clip 覆盖数据行内容（P0）

**影响屏**：开奖查询

**根因**：`drawTableRow` 中使用 `defer clip.Rect{...}.Push(gtx.Ops).Pop()` 为背景填充创建裁剪区域。由于 `defer` 在整个函数返回前不释放，该 clip 一直活跃到分割线渲染阶段。分割线的 `paint.Fill` 本意是画 1px 细线，但活跃的 clip 仍是整行矩形，导致分割线颜色覆盖了整行文字内容。

**时序图**：
```
drawTableRow 执行流:
  defer clip.Push  ← 建立整行矩形 clip
  paint.Fill(背景)  ← ✅ 背景填充正确
  Flex → 文本渲染   ← ✅ 文字叠在背景上
  Flex → 分割线     ← ❌ paint.Fill(分割线色) 覆盖整个 clip → 文字被盖
  defer Pop        ← 太晚了
```

**修复**：每次 `paint.Fill` 前后独立管理 clip 生命周期：

```go
// 背景 — 立即 Pop
bgClip := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
paint.Fill(gtx.Ops, bg)
bgClip.Pop()

// ... 文本渲染 ...

// 分割线 — 独立 clip，约束已设为 Max.Y = 1dp
divClip := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
paint.Fill(gtx.Ops, th.Colors.TableDivider)
divClip.Pop()
```

### 3.3 goroutine panic 静默丢失（P1）

**影响屏**：开奖查询

**根因**：`fetchDrawsPageAsync` 在 goroutine 中执行，无 `recover()` 保护。若发生 panic：
- `defer` 中 `state.Loading = false` 仍执行
- 但 `state.Draws` 未赋值（始终为 nil）
- `state.Error` 未设置（始终为 nil）
- UI 落入了 `len(state.Draws) == 0` 分支 → 显示"暂无开奖数据"，无任何错误提示

**修复**：在 defer 中增加 panic 捕获，转为 `state.Error`：

```go
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
```

### 3.4 错误状态未重置（P2）

**影响屏**：开奖查询

**根因**：重新拉取数据（翻页/刷新）时未清除上一次的 `state.Error`，可能导致 error UI 残留。

**修复**：所有触发数据拉取的路径均添加 `state.Error = nil`：
- 首次自动加载
- 刷新按钮点击
- 上一页/下一页点击

---

## 4. 渲染模式精简

### 变更

`drawTableHeader` 和 `drawTableRow` 从 `layout.Stack{Expanded + Stacked}` 双通道模式改为更简单的 `clip.Rect` + `paint.Fill` 单通道模式。

**原因**：`layout.List` 对其子元素尺寸计算有严格要求。Stack 的 `Expanded` 子元素返回 `layout.Dimensions{}`（零尺寸），虽然理论上 Stack 取自 `Stacked` 尺寸，但在 gioui 的 List 引擎中可能导致尺寸传播异常。项目原始可工作的卡片布局（`drawCard`）同样使用 `clip.Rect` 模式。

### 对比

```go
// 旧：Stack 双通道（已移除）
layout.Stack{Alignment: layout.NW}.Layout(gtx,
    layout.Expanded(func(gtx layout.Context) layout.Dimensions {
        paint.Fill(gtx.Ops, bg)
        return layout.Dimensions{}  // 零尺寸 → List 可能误判
    }),
    layout.Stacked(func(gtx layout.Context) layout.Dimensions {
        // 实际内容
    }),
)

// 新：clip.Rect 单通道
defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
paint.Fill(gtx.Ops, bg)
// 直接渲染内容 → 尺寸由内容决定
```

---

## 5. 修改文件汇总

| 文件 | 变更类型 | 说明 |
|---|---|---|
| `ui/theme/palette.go` | 新增 | `TableHeaderBg`, `TableAltBg`, `TableDivider` 三个 Token |
| `ui/screen/history.go` | 重构+Bug修复 | 消除硬编码色、Axis 修复、clip 修复、panic 保护、错误重置 |
| `ui/screen/stats.go` | Bug修复 | Axis 修复 |

---

## 6. 渲染问题排查历程

| 步骤 | 排查内容 | 结论 |
|---|---|---|
| 1 | 验证 API 端点可达性 | ✅ `curl` 返回 200，JSON 结构正确 |
| 2 | 验证数据解析链路 | ✅ `go run` 诊断脚本：30 条/页，total=2894 |
| 3 | 验证 service → client 调用链 | ✅ 端到端 Go 测试通过 |
| 4 | 对比原始 `drawCard`（git HEAD）与当前 `drawTableRow` | 发现 Stack 模式差异 |
| 5 | 检查 `layout.List.Axis` 默认值 | 🔴 零值 = Horizontal，根因 #1 |
| 6 | 对比 history vs stats 表现（Axis 修复后 stats 正常） | history 仍有问题，定位到 `drawTableRow` |
| 7 | 分析 `defer clip` 生命周期 | 🔴 分割线 Fill 覆盖文字，根因 #2 |

---

## 7. 验证结果

- [x] `gofmt` 格式化通过
- [x] `go build .` 编译通过
- [x] `go test ./...` 全部通过
- [x] 表头正常渲染
- [x] 数据行正常渲染
- [x] 分页控件正常交互
- [x] 冷热统计页面正常
