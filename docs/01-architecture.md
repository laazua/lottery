# 01-architecture.md — 软件架构文档

## 1. 项目概述

### 1.1 项目定位

本 APP 是一款手机移动端大乐透（超级大乐透）往期数据查询、冷热号分析以及智能推荐选号的工具型应用。目标用户为购买大乐透彩票的彩民群体，提供便捷的历史数据查阅和科学的选号参考。

### 1.2 核心功能

| 功能 | 描述 | 优先级 |
|------|------|--------|
| 历史开奖查询 | 从 cwl.gov.cn 公开 API 拉取并展示往期开奖结果，支持期号搜索 | P0 |
| 冷热号统计 | 对前区（1-35）和后区（1-12）号码进行频次统计，分为冷/温/热三档 | P0 |
| 智能推荐 | 基于冷热统计、遗漏值等维度给出推荐号码组合 | P1 |
| 号码详情 | 查看特定号码的历史出现分布 | P2 |

## 2. 技术选型

| 层级 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 语言 | Go | 1.26.1 | 主编程语言 |
| UI 框架 | gioui | 0.10.0 | 跨平台即时模式 GUI 框架 |
| 数据源 | cwl.gov.cn 公开 API | — | 中国福利彩票发行管理中心公开接口 |
| 日志 | log/slog | 标准库 | Go 1.21+ 结构化日志 |
| HTTP | net/http | 标准库 | HTTP 客户端 |
| JSON | encoding/json | 标准库 | JSON 序列化/反序列化 |

## 3. 整体架构

### 3.1 分层架构图

```
┌────────────────────────────────────────────────────────────┐
│                        main.go                              │
│              入口：Window 初始化 + 依赖组装                    │
└────────────────────┬───────────────────────────────────────┘
                     │ 注入 client 实现
                     ▼
┌────────────────────────────────────────────────────────────┐
│                        ui/ 层                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────┐  │
│  │ app.go   │→│ router.go│→│ screen/  │→│ widget/   │  │
│  │窗口/主题  │  │ 导航路由  │  │ 业务屏   │  │ 可复用组件│  │
│  └──────────┘  └──────────┘  └────┬─────┘  └───────────┘  │
└───────────────────────────────────┼─────────────────────────┘
                                    │ 调用 service 接口
                                    ▼
┌────────────────────────────────────────────────────────────┐
│                      service/ 层                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ lottery.go   │  │ statistics.go│  │recommendation.go │  │
│  │ 开奖数据拉取  │  │ 冷热号统计   │  │ 推荐算法         │  │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘  │
└─────────┼──────────────────┼───────────────────┼────────────┘
          │                  │                   │
          ▼                  ▼                   ▼
┌────────────────────────────────────────────────────────────┐
│                      model/ 层                               │
│       ┌──────────┐  ┌──────────┐  ┌────────────────┐       │
│       │ draw.go  │  │statistics│  │recommendation  │       │
│       │开奖数据   │  │ 统计结果  │  │ 推荐结果        │       │
│       └──────────┘  └──────────┘  └────────────────┘       │
│                       ▲ 依赖接口，不依赖实现                    │
└───────────────────────┼────────────────────────────────────┘
                        │
                        ▼
┌────────────────────────────────────────────────────────────┐
│                       client/ 层                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ cwl.go — cwl.gov.cn API 对接实现                      │  │
│  │ 职责：HTTP 请求、签名、限流、响应解析、错误映射          │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────┘
```

### 3.2 依赖方向规则

```
main.go → ui/ → service/ → model/
                           → client/ (接口)
                                  → cwl.go (实现)
              internal/→errors/ (工具函数，无上层依赖)
                       →period/
                       →httputil/
```

**核心约束**：
- `service/` 不可反向依赖 `ui/`
- `model/` 不可依赖任何其他包
- `client/` 只依赖标准库和 `model/`
- 所有外部依赖必须通过 `client/` 抽象层封装

## 4. 包职责矩阵

| 包路径 | 职责 | 对外暴露 | 依赖 |
|--------|------|---------|------|
| `client/` | 封装 cwl.gov.cn API 调用，提供统一的 LotteryAPI 接口 | `LotteryAPI` 接口 + 实现 | 标准库, model |
| `model/` | 纯数据定义，无方法逻辑 | `Draw`, `Statistics`, `Recommendation` 等 struct | 无 |
| `service/` | 核心业务逻辑：数据组装、统计分析、推荐算法 | `LotteryService`, `StatsService`, `RecommendService` | model, client(接口) |
| `ui/` | gioui 窗口、主题、导航、屏幕组件 | `App`, `Router`, 各 Screen | service, model, gioui |
| `ui/theme/` | 调色板、字体、动画参数 | `Theme` struct 和配色常量 | gioui |
| `ui/screen/` | 功能屏编排：组合 widget 成完整页面 | 每个屏的 Layout 函数 | service, widget, theme |
| `ui/widget/` | 可复用 UI 组件：号码球、条形图、卡片 | 组件函数和状态 | theme, model, gioui |
| `internal/errors/` | 自定义错误类型、错误码定义 | 本仓库内部 | 标准库 |
| `internal/period/` | 期号解析与计算工具（期号格式校验、年月提取） | 本仓库内部 | 标准库 |
| `internal/httputil/` | HTTP 重试、超时、限流等通用工具 | 本仓库内部 | 标准库 |

## 5. 数据流

### 5.1 历史数据查询流程

```
用户进入首页
    │
    ▼
screen/history.go → goroutine 启动 LoadDraws(ctx, page)
    │                                  │
    │  (当前帧继续渲染 loading 状态)     │
    │                                  ▼
    │                     service/lottery.go 调用 client.FetchDraws(ctx, params)
    │                                  │
    │                                  ▼
    │                     client/cwl.go 构造 HTTP 请求 → GET cwl.gov.cn API
    │                                  │
    │                                  ▼
    │                     HTTP 响应 → JSON 解析 → model.Draw 列表
    │                                  │
    │  ← ← ← ← ← ← ← ← ← ← ← ← ← ← 返回数据
    │
    ▼
更新 HistoryState → w.Invalidate() 触发重绘
    │
    ▼
下一帧 → screen/history.go 读取更新后的 state → 渲染号码列表
    │
    ▼
widget/ball.go 逐号码渲染
```

> ⚠️ **gioui 异步关键模式**：所有网络请求必须在 goroutine 中执行，禁止在 Layout 方法中直接调用阻塞 IO。
> 完成回调中调用 `w.Invalidate()` 触发下一帧重绘，screen 在下一帧读到更新后的 state 渲染结果。

### 5.2 冷热统计流程

```
用户选择统计期数范围（20/50/100期）
    │
    ▼
screen/stats.go → goroutine 启动 service.CalculateStats(ctx, draws, range)
    │
    ▼
service/statistics.go：
    1. 遍历指定范围内所有开奖数据
    2. 统计前区 1-35 每个号码出现频次
    3. 统计后区 1-12 每个号码出现频次
    4. 按频次降序排列
    5. 前 30% = 热号，后 30% = 冷号，中间 = 温号
    │
    ▼
返回 model.Statistics → 更新 state → w.Invalidate() 触发重绘
    │
    ▼
screen/stats.go 读取更新后的 state → 渲染冷温热分区
```

### 5.3 推荐号码流程

```
用户点击"生成推荐"
    │
    ▼
screen/recommend.go → goroutine 启动 service.GenerateRecommendation(ctx, draws)
    │
    ▼
service/recommendation.go：
    1. 获取最新统计数据（频次排序 + 遗漏值排行）
    2. 从前区热号池中加权随机选 3 个  ← 前区已选 3
    3. 从前区温号池中加权随机选 1 个  ← 前区已选 4
    4. 从前区遗漏值最高的号码池中选 1 个  ← 前区已选 5 ✅
    5. 从后区（热号+温号）合并池中加权随机选 1 个  ← 后区已选 1
    6. 从后区遗漏值最高的号码池中选 1 个  ← 后区已选 2 ✅
    7. 验证最终组合无重复号码、号码数量正确（前区5+后区2）、升序排列
    │
    ▼
返回 model.Recommendation → 更新 state → w.Invalidate() 触发重绘
    │
    ▼
screen/recommend.go 展示推荐号码（每个号码带推荐理由标签：🔥热/🌡️温/📊遗漏）
```

### 5.4 gioui 异步更新模式说明

gioui 是即时模式（immediate mode）GUI 框架，每帧全量重绘。在此架构中，异步数据更新的标准模式如下：

```
帧循环（~60fps）
    │
    ├── FrameEvent 到达
    │       │
    │       ▼
    │   screen.Layout(gtx, state)
    │       │
    │       ├── 读取 appState
    │       ├── 根据 state.Loading 分支渲染：
    │       │   ├── true  → 渲染 loading 骨架屏
    │       │   └── false → 渲染数据视图
    │       │
    │       └── 若用户触发了操作（点击"加载"等）
    │               │
    │               └── 启动 goroutine:
    │                       ├── 执行 service.Operation(ctx)
    │                       ├── 等待完成或超时/出错
    │                       └── appState 更新 → w.Invalidate()
    │
    └── End of frame（等待下一个 FrameEvent）
```

**禁止模式**：在 Layout 方法内直接调用任意阻塞 IO（HTTP 请求、文件读写等）。

### 5.5 gioui 事件处理关键规则

gioui 是即时模式框架，与传统的保留模式（Retained Mode）GUI 框架有本质区别。以下规则是代码能正常交互的前提：

**规则一：widget 状态必须跨帧持久化**

```go
// ✅ 正确：Clickable 存储在 State 中，每帧复用同一实例
type HistoryState struct {
    LoadButton widget.Clickable   // ← 存储在 State 中
    TabButtons [3]widget.Clickable
    Draws      []model.Draw
}

// ❌ 错误：每帧新建 Clickable，点击事件永远丢失
func BadLayout(gtx layout.Context) {
    btn := new(widget.Clickable)  // ← 每帧新建，上一帧的状态丢失
    material.Button(btn, "点击").Layout(gtx)
}
```

**规则二：事件回调在 Layout 中通过 Clickable.Clicked(gtx) 检测**

```go
// ✅ 正确模式
func HistoryLayout(gtx layout.Context, state *HistoryState, route *Route) {
    if state.LoadButton.Clicked(gtx) {
        route.Current = ScreenRecommend  // 更新路由
    }
    material.Button(&state.LoadButton, "加载").Layout(gtx)
}
```

**规则三：数据加载触发器在 Layout 顶部检测**

```go
// ✅ 首次进入屏时自动加载数据
func HistoryLayout(gtx layout.Context, state *HistoryState) {
    if !state.Loaded && !state.Loading {
        state.Loading = true
        go fetchDataAsync(state)  // goroutine 加载 → 回调中 w.Invalidate()
    }
    // ... 渲染
}
```

**规则四：所有可交互 widget（Clickable、Editor、Slider 等）都必须做持久化存储**。常见遗忘点：

| 组件 | 存储位置 | 常见错误 |
|------|---------|---------|
| `widget.Clickable` | State struct 字段 | 在 Layout 中用 `new(Clickable)` |
| `widget.Editor` | State struct 字段 | 每次 Layout 重新创建 |
| `widget.Float` | State struct 字段 | 未持久化导致手势无效 |
| `layout.List` | State struct 字段 | 丢失后滚动位置重置 |

## 6. UI 三屏设计规格

整体设计语言采用 **Material Design 3（Material You）风格**，核心特征是圆角卡片、层次化阴影、克制的用色和呼吸感间距。

### 6.1 视觉 Token

| Token | 值 | 用途 |
|-------|-----|------|
| 背景色 | `#F5F5F5`（浅灰） | 整体页面背景 |
| 卡片色 | `#FFFFFF`（白） | 开奖条目、统计卡片 |
| 主色 | `#E53935`（红） | 后区号码球、选中态 |
| 辅色 | `#1E88E5`（蓝） | 前区号码球、链接 |
| 热号色 | `#E53935`（红） | 冷热统计热号标记 |
| 温号色 | `#FB8C00`（橙） | 冷热统计温号标记 |
| 冷号色 | `#2196F3`（蓝） | 冷热统计冷号标记 |
| 号球文字 | `#FFFFFF`（白） | 号码球上数字 |
| 按钮圆角 | 20dp（胶囊形） | 所有按钮统一胶囊造型 |
| 卡片圆角 | 10dp | 略小号圆角更紧凑 |
| 号球圆角 | 50%（圆形） | 完整圆形 |
| 卡片阴影 | elevation 2dp | 层次感 |
| 间距层级 | 2/4/6/8/12/16/24dp | 紧凑布局，减少留白 |

### 6.2 底部导航栏

```
┌────────────────────────────────────────────┐
│   📋 开奖查询    📊 冷热统计    🎯 智能推荐 │  ← 图标 + 文字
│        ● 选中态指示条                      │  ← 当前屏底部有圆点
└────────────────────────────────────────────┘
```

- 三个 Tab 平级切换，无栈式导航
- 选中 Tab 图标/文字用主色，非选中用灰色
- 选中 Tab 底部显示 8dp 宽圆角指示条
- 切换时内容区域无过渡动画（gioui 限制，纯色切换）

### 6.3 历史数据屏（History）

- **布局**：全屏列表 + 底部 Tab 栏
- **顶部**：标题 "开奖查询" + 右侧刷新按钮
- **列表条目**（圆角白色卡片，间距 8dp）：
  ```
  ┌─────────────────────────────────┐
  │ 第 24180 期    2026-07-04      │ ← 期号 + 日期
  │ ┌──────────────┐ ┌──┐         │
  │ │05│12│18│23│31│ │07│11│      │ ← 前区蓝色球 + 后区红色球
  │ └──────────────┘ └──┘         │
  │ 销售额: 3.1亿    奖池: 9.2亿   │ ← 灰色辅助信息
  └─────────────────────────────────┘
  ```
- **功能**：首次进入自动拉取最近 20 期、向下滚动追加加载、期号搜索
- **空状态**："暂无开奖数据" 居中占位 + 灰色图标

### 6.6 按钮组件样式标准

所有按钮统一使用 `widget/button.go` 提供的样式化组件，消除 `material.Button` 的默认外观。

| 按钮类型 | 用途 | 样式参数 |
|---------|------|---------|
| `FilledBtn` | 主要操作（刷新、生成推荐） | 主色填充、白色文字、胶囊圆角（20dp） |
| `OutlineBtn` | 次要操作（期数切换非选中态） | 灰色描边、深色文字、胶囊圆角 |
| `SmallBtn` | 紧凑场景（期数切换选中态） | 主色填充、缩小内边距（6dp/12dp） |

```go
// 使用示例
btn := lotwidget.FilledBtn(th, "刷新", &state.RefreshBtn)
btn.Layout(gtx)
```

### 6.4 冷热统计屏（Stats）

- **布局**：纵向滚动内容 + 底部 Tab 栏
- **顶部**：统计窗口选择器 [近20期] [近50期] [近100期]
- **前区分区**（标题 + 三行冷温热号球）：
  ```
  🔥 热号    07 12 18 23 25 31 33 05   ← 红色号码球
  🌡️ 温号    01 04 09 14 17 20 22 27  ← 橙色号码球
  ❄️ 冷号    02 03 06 08 10 11 15 16  ← 蓝色号码球
  ```
- **后区分区**：同上，展示 1-12 号
- **频率条**：水平条形图，按频次降序排，渐变色填充
- **功能**：切换统计范围时重新计算冷温热分档（30%/40%/30%）

### 6.5 推荐选号屏（Recommend）

- **布局**：居中内容 + 底部 Tab 栏
- **顶部**：标题 "智能推荐" + "生成推荐" 按钮（带渐变背景）
- **推荐号码展示**：
  ```
  🎯 前区推荐
  ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐
  │ 07 │ │ 12 │ │ 18 │ │ 23 │ │ 31 │  ← 大号球，下方标注推荐理由
  │🔥热│ │🔥热│ │🔥热│ │🌡️温│ │📊遗漏│
  └────┘ └────┘ └────┘ └────┘ └────┘

  🎯 后区推荐
  ┌────┐ ┌────┐
  │ 07 │ │ 11 │
  │🔥热│ │📊遗漏│
  └────┘ └────┘
  ```
- **功能**：点击"生成推荐"调用推荐算法、结果展示每个号码的推荐理由标签
- **未生成时状态**："点击生成推荐号码" 居中占位提示

## 7. 冷热判定算法

### 7.1 算法选择

采用 **排序百分比分档法**，而非标准差阈值法。原因：彩票开奖号码在大样本下趋近均匀分布，标准差法在均匀分布数据中区分度低，易使大部分号码归入"温号"。

### 7.2 判定规则

```
1. 对前区 1-35 各号码按出现频次降序排列
2. 前 30%（取整，约 10 个号码）→ 🔥 热号
3. 后 30%（取整，约 10 个号码）→ ❄️ 冷号
4. 中间 40%（约 15 个号码）→ 🌡️ 温号
5. 后区 1-12 同理：前 30%→热(约4个)，后 30%→冷(约4个)，中间→温(约4个)
```

### 7.3 遗漏值计算

遗漏值 = 当前最近一期到该号码上次出现之间的期数差。

遗漏值用于：
- 辅助推荐算法选择冷号（高遗漏值号码有回补概率）
- 在统计屏中展示"最大遗漏"维度的信息

## 7.4 构建时配置体系

数据源 API 地址通过 **构建时注入（ldflags）** 配置，不硬编码在代码中。允许同一份代码在不同构建中指向不同数据源。

### 配置变量

| 变量 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `APIBaseURL` | string | `https://www.cwl.gov.cn` | 数据源 API 基础 URL |
| `DataSource` | string | `cwl` | 数据源类型（`cwl`/`mock`） |

### 注入方式

```bash
# 注入 API 地址（桌面调试）
go run -ldflags="-X 'github.com/user/lottery/internal/config.APIBaseURL=https://api.example.com'" .

# 注入数据源类型（使用模拟数据）
go run -ldflags="-X 'github.com/user/lottery/internal/config.DataSource=mock'" .
```

### 代码结构

```go
// internal/config/config.go
package config

var (
    APIBaseURL = "https://www.cwl.gov.cn"  // 默认福彩官网
    DataSource = "cwl"                      // cwl=在线 | mock=模拟
)
```

通过 Makefile 目标统一管理（见 `09-packaging-deployment.md`）。

## 8. 导航与状态管理

### 8.1 导航策略

使用声明式屏幕路由。`router.go` 维护当前屏幕标识符和参数：

```go
type ScreenID int

const (
    ScreenHistory   ScreenID = iota
    ScreenStats
    ScreenRecommend
)

type Route struct {
    Current   ScreenID
    Params    map[string]any  // 跨屏参数传递
}
```

导航方式：底部 Tab 栏三屏切换，无需栈式导航（平级关系）。

**路由更新逻辑**：底部 Tab 栏的每个 `Clickable` 检测到点击后，更新 `route.Current`，下一帧自动渲染新屏。

### 8.2 State 结构体设计

每个屏的 State 结构体必须包含该屏所有可交互组件的持久化状态。这是 gioui 框架的硬性要求：

```go
// HistoryState 包含历史屏的 UI 数据和所有交互组件状态。
type HistoryState struct {
    // 业务数据
    Draws      []model.Draw
    Loading    bool
    Loaded     bool      // 是否已加载过（用于首次自动加载判断）
    Error      error

    // gioui 交互组件（必须持久化，禁止在 Layout 中新建）
    RefreshBtn widget.Clickable   // 刷新按钮
    List       layout.List       // 可滚动列表（保持滚动位置）
    SearchEdit widget.Editor     // 搜索输入框

    // 搜索防抖
    searchTerm string
    searchDebounce *time.Timer
}
```

关键规则：
1. **所有 `widget.Clickable`、`widget.Editor`、`layout.List` 等有状态的组件必须放在 state 中**
2. **禁止在 Layout 方法中 `new(Clickable)`** — 这会导致每帧新建，点击永远丢失
3. **`layout.List` 也需持久化** — 否则滚动位置每帧重置

### 8.3 首次加载模式（OnAppear）

每个屏在首次展示时自动触发数据加载。模式如下：

```go
func HistoryLayout(gtx layout.Context, state *HistoryState, svc *Services) {
    // 首次进入且未开始加载 → 触发异步加载
    if !state.Loaded && !state.Loading {
        state.Loading = true
        state.Loaded = true  // 防止重复触发
        go func() {
            draws, err := svc.Lottery.FetchDraws(context.Background(), 20)
            // 注意：需要在 goroutine 中通过 w.Invalidate() 触发重绘
            // 实际代码中需注入 Window 引用或使用回调
            state.Draws = draws
            state.Error = err
            state.Loading = false
            // w.Invalidate()  → 下一帧重绘
        }()
    }

    // 后续渲染逻辑...
}
```

### 8.4 Tab 切换事件流

```
用户点击底部"冷热统计"Tab
    │
    ▼
检测 state.TabBtn[ScreenStats].Clicked(gtx) == true
    │
    ▼
route.Current = ScreenStats
    │
    ▼
下一帧 App.Layout() 根据 route.Current 路由到 StatsLayout()
    │
    ▼
StatsLayout() 检测到 !state.Loaded → 触发统计数据加载
```

### 8.5 错误状态处理

所有 API 调用均返回 `error`，UI 层统一处理为：
1. 网络错误 → Toast 提示 + 重试按钮
2. 解析错误 → slog 记录 + 数据异常提示
3. 空数据 → 空状态占位图

## 9. 移动端生命周期适配

gioui 在移动端（Android）运行时涉及以下生命周期事件：

| 事件 | 触发时机 | 处理策略 |
|------|---------|---------|
| `app.ConfigEvent` | 屏幕旋转、窗口大小变化 | 记录新配置，下次 Layout 自动适配 |
| `app.StageEvent` (Pause) | APP 切到后台 | 若正在进行的网络请求，标记 context 取消 |
| `app.StageEvent` (Resume) | APP 从后台恢复 | 重新发起被取消的请求（若有） |
| `app.ViewEvent` | 原生 View 创建/更新 | gioui 内部管理，应用层无需处理 |
