# 02-coding-guidelines.md — 编码规范

## 1. 命名规范

### 1.1 包命名

| 规则 | 示例 | 禁止 |
|------|------|------|
| 小写字母，单数形式 | `client`, `service`, `model` | `clients`, `myPackage` |
| 与目录名一致 | 目录 `service/` → 包名 `service` | 目录 `stats/` 但包名 `statistic` |
| 不使用下划线 | `httputil` | `http_util` |
| 不使用复数 | `draw` | `draws` |

### 1.2 类型命名

- 导出类型：首字母大写 + 驼峰，如 `DrawResult`, `HotColdStats`
- 非导出类型：首字母小写 + 驼峰，如 `httpClient`, `statsCalculator`
- 接口名：以 "er" 后缀或行为描述结尾，如 `LotteryAPI`, `StatsProvider`

### 1.3 函数与方法命名

- 导出函数：首字母大写，动词开头，如 `FetchDraws`, `CalculateStats`
- 非导出函数：首字母小写，如 `parseDrawResponse`, `validateParams`
- Getter：Go 中省略 Get 前缀，如 `Draws()` 而非 `GetDraws()`
- Setter：使用 Set 前缀，如 `SetRange(n int)`

### 1.4 常量与变量

- 常量：驼峰，导出大写首字母，如 `DefaultPageSize = 20`
- 变量：驼峰，短名（作用域越小名越短）
- 错误变量：以 `Err` 开头，如 `ErrInvalidPeriod`, `ErrNetworkFailure`

## 2. 文件组织规范

### 2.1 文件命名

| 规则 | 示例 |
|------|------|
| 小写字母 + 下划线分词 | `draw_result.go`, `http_client.go` |
| 测试文件以 `_test.go` 结尾 | `statistics_test.go`, `cwl_client_test.go` |
| 一个包的入口文件与包同名 | `client/client.go` 定义接口 |

### 2.2 文件内组织顺序

```
1. 包声明 (package xxxx)
2. 导入 (import)
3. 常量定义
4. 类型定义
5. 变量声明
6. 构造函数 (NewXxx)
7. 导出方法
8. 非导出方法
9. 辅助类型和函数
```

### 2.3 文件大小建议

- 单个文件不超过 300 行
- 超过时考虑拆分子文件，如 `cwl_client.go` + `cwl_parser.go` + `cwl_errors.go`

## 3. 注释规范

### 3.1 导出名称注释

所有导出的包、类型、函数、方法必须有文档注释，格式为：

```go
// Package client 提供外部数据源访问抽象层。
package client

// DrawResult 表示大乐透单期开奖结果。
type DrawResult struct { ... }

// FetchDraws 从公开 API 拉取指定范围的开奖数据。
// 返回 DrawResult 列表；当网络不可用时返回 ErrNetworkFailure。
func FetchDraws(ctx context.Context, opts ...Option) ([]DrawResult, error)
```

注释规则：
- 以被注释的名称**开头**
- 以**句号结尾**
- 完整句子，非短语
- 可包含使用示例

### 3.2 内部代码注释

```go
// 非导出函数的注释，首字母可不与名称一致（非强制）

// 复杂逻辑的关键步骤
// Step 1: 验证期号格式
// Step 2: 解析期号中的年月信息
```

### 3.3 TODO/FIXME 注释

```go
// TODO(zhangsan): 后续接入更多数据源时需抽取为策略模式
// FIXME: 边缘情况：当期号为"00000"时会 panic
```

## 4. 错误处理规范

### 4.1 基本原则

错误处理必须显式检查，禁止用 `_` 丢弃 error：

```go
// ✅ 正确
resp, err := client.FetchDraws(ctx, params)
if err != nil {
    return fmt.Errorf("拉取开奖数据失败: %w", err)
}

// ❌ 错误
resp, _ := client.FetchDraws(ctx, params)
```

### 4.2 错误包装

使用 `fmt.Errorf` + `%w` 创建错误链：

```go
// 下层
var ErrInvalidPeriod = errors.New("无效期号格式")

// 中间层
if err := validatePeriod(p); err != nil {
    return fmt.Errorf("查询开奖数据: %w", err)
}

// 上层（UI 层展示）
slog.Error("开奖数据查询失败", "period", p, "error", err)
ui.ShowToast("数据加载失败，请稍后重试")
```

### 4.3 错误类型

| 类型 | 适用场景 | 处理方式 |
|------|---------|---------|
| 哨兵错误 (var ErrXxx) | 可预期的业务异常 | `errors.Is` 判断 |
| 自定义错误类型 | 需要额外上下文 | 实现 `Error()` 方法 |
| 包装错误 (%w) | 调用链传递 | 上层逐层解包判断 |

### 4.4 panic 使用限制

- **禁止**用 panic 处理普通错误
- **允许**：程序初始化阶段不可恢复的异常
- **允许**：gioui 的 goroutine 内部不可恢复状态

## 5. 日志规范

### 5.1 使用 slog

全项目统一使用 `log/slog`，禁止使用 `fmt.Println`：

```go
// ✅ 正确
slog.Info("开始拉取开奖数据",
    "period", period,
    "page", page,
)

// ❌ 错误
fmt.Println("拉取数据中...")
```

### 5.2 日志级别

| 级别 | 使用场景 |
|------|---------|
| `Debug` | 开发调试细节，生产环境默认关闭 |
| `Info` | 关键业务事件（数据拉取成功、推荐生成完成） |
| `Warn` | 可恢复的异常（API 限流回退、重试首次失败） |
| `Error` | 不可恢复的错误（API 连续失败、解析出错） |

### 5.3 日志格式

```go
slog.Error("拉取开奖数据失败",
    "source", "cwl.gov.cn",
    "period", period,
    "attempt", retryCount,
    "error", err,
)
```

## 6. gioui 编码约定

### 6.1 核心规则：widget 必须跨帧持久化（硬性）

gioui 是即时模式框架，每帧全量重绘。所有有状态的 widget（`Clickable`、`Editor`、`List`、`Float` 等）**必须在 State 结构体中声明**，保证帧间复用同一个实例。

```go
// ✅ 正确：Clickable 存储在 State 中
type MyState struct {
    Btn widget.Clickable   // 持续跟踪点击状态
    Edt widget.Editor      // 持续跟踪输入内容和光标
    Lst layout.List        // 持续跟踪滚动位置
}

func MyLayout(gtx layout.Context, state *MyState) layout.Dimensions {
    if state.Btn.Clicked(gtx) {
        // 处理点击事件
    }
    material.Button(&state.Btn, "点击").Layout(gtx)
    //        ↑ 传指针，每帧用同一个 Clickable
}
```

```go
// ❌ 错误：Clickable 在 Layout 中新建，帧间状态丢失
func WrongLayout(gtx layout.Context) layout.Dimensions {
    btn := new(widget.Clickable)  // 帧帧新建，点击事件永远不触发
    material.Button(btn, "点击").Layout(gtx)
}
```

### 6.2 布局与事件处理分离

- Layout 方法中**检测**事件（`Clicked(gtx)`），但不包含耗时逻辑
- 耗时操作在 goroutine 中执行，完成后通过 `w.Invalidate()` 触发重绘

```go
// ✅ 正确模式
func HistoryLayout(gtx layout.Context, th *theme.Theme, state *HistoryState) {
    // ① 检测事件（顶部）
    if state.RefreshBtn.Clicked(gtx) {
        go fetchAndUpdate(state)  // goroutine 异步加载
    }

    // ② 绘制 UI（底部）
    material.Button(&state.RefreshBtn, "刷新").Layout(gtx)
}

func fetchAndUpdate(state *HistoryState) {
    state.Loading = true
    draws, err := lotterySvc.FetchDraws(context.Background(), 20)
    state.Draws = draws
    state.Error = err
    state.Loading = false
    // 调用 w.Invalidate() 触发重绘（状态已更新，下一帧渲染）
}
```

### 6.3 Tab 导航模式

底部 Tab 栏的每个 Tab 必须有一个持久化的 `Clickable`，导航事件在 Layout 顶部检测：

```go
// state 中定义
type AppState struct {
    TabBtns [3]widget.Clickable  // 每个 Tab 一个 Clickable
    Route   screen.Route
}

// Layout 中检测
func BottomTabLayout(gtx layout.Context, state *AppState) {
    for i := range state.TabBtns {
        if state.TabBtns[i].Clicked(gtx) {
            state.Route.Current = screen.ScreenID(i)
        }
    }
    // ... 渲染 Tab 按钮
}
```

### 6.4 首次加载模式

每个屏在首次渲染时自动触发数据加载：

```go
// 在 Layout 方法顶部（布局代码之前）
if !state.Loaded && !state.Loading {
    state.Loading = true
    state.Loaded = true
    go func() {
        draws, err := svc.Lottery.FetchDraws(context.Background(), 20)
        state.Draws = draws
        state.Error = err
        state.Loading = false
        // w.Invalidate() → 下一帧重绘
    }()
}
```

### 6.5 布局与状态分离

- Layout 方法仅负责框架嵌套和组件排列
- 所有可变状态通过 `*State` 指针传入，不在 Layout 内部定义

```go
// ✅ 推荐
func HistoryLayout(gtx layout.Context, th *theme.Theme, state *HistoryState) layout.Dimensions {
    // 仅布局编排 + 事件检测
}

// ❌ 不推荐：Layout 方法超过 80 行
// 拆分为 Header / List / BottomTab 等子函数
```

### 6.6 主题统一

- 颜色、字体、间距等设计 Token 统一定义在 `ui/theme/palette.go`
- widget 组件优先使用 theme 中的值，而非硬编码
- 所有颜色值使用 `color.NRGBA`，避免 `color.RGBA`（不兼容 gioui）
