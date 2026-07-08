# 04-api-version-specification.md — API 版本规范

## 1. 对外数据源 API

### 1.1 数据源说明

本 APP 使用中国福利彩票发行管理中心公开的开奖数据接口：

| 项目 | 内容 |
|------|------|
| 数据源 | cwl.gov.cn（中国福利彩票发行管理中心） |
| 数据内容 | 超级大乐透（大乐透）历史开奖号码 |
| 请求方式 | HTTP GET |
| 数据格式 | JSON |
| 限流要求 | 建议请求间隔 ≥1 秒，避免触发服务端限流 |

### 1.2 API 端点

#### 查询历史开奖数据

> ⚠️ **实现前核实**：cwl.gov.cn 的接口模式可能随站点改版变化，以下端点模式需在实现阶段通过浏览器开发者工具抓包确认。
>
> 已知参考模式：
> - 双色球：`https://www.cwl.gov.cn/cwl/f1/ssq/`（仅参考，大乐透类似）
> - 大乐透可能使用独立子域名或路径
>
> 实现时建议使用版本化 HTML 页面解析（如 `https://www.cwl.gov.cn/cwl/f1/xxx/`）或 JSON 接口。以下为预期接口规范，最终以实测为准。

```
GET [待实现阶段核实的 URL]
```

**请求参数（预期）**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 彩种名称，如 "超级大乐透" |
| `issueCount` | int | 否 | 查询期数，默认 20 |
| `startIssue` | string | 否 | 起始期号 |
| `endIssue` | string | 否 | 截止期号 |

**响应格式（预期）**：

```json
{
    "total": 100,
    "data": [
        {
            "issue": "24180",
            "drawTime": "2026-07-04 21:20:00",
            "frontWinningNum": "05,12,18,23,31",
            "backWinningNum": "07,11",
            "saleAmount": "310000000",
            "poolAmount": "920000000"
        }
    ]
}
```

### 1.3 DTO 映射

```go
// model/draw.go

// Draw 表示单期大乐透开奖结果。
type Draw struct {
    Issue         string    // 期号，如 "24180"
    DrawTime      time.Time // 开奖日期
    FrontNumbers  [5]int    // 前区号码（1-35），升序排列
    BackNumbers   [2]int    // 后区号码（1-12），升序排列
    SaleAmount    int64     // 销售额（元）
    PoolAmount    int64     // 奖池金额（元）
}
```

### 1.4 DTO 转换

`client/cwl.go` 中完成 JSON 响应到 `model.Draw` 的转换：

```go
type cwlResponse struct {
    Total int          `json:"total"`
    Data  []cwlDrawDTO `json:"data"`
}

type cwlDrawDTO struct {
    Issue         string `json:"issue"`
    DrawTime      string `json:"drawTime"`
    FrontWinningNum string `json:"frontWinningNum"` // "05,12,18,23,31"
    BackWinningNum  string `json:"backWinningNum"`  // "07,11"
    SaleAmount    string `json:"saleAmount"`
    PoolAmount    string `json:"poolAmount"`
}
```

## 2. 接口版本管理

### 2.1 cwl.gov.cn 版本兼容策略

| 变更类型 | 应对策略 | 处理时间要求 |
|---------|---------|-------------|
| 响应字段新增 | 自动兼容，忽略未知字段 | 不追溯 |
| 响应字段删除 | client 层适配，添加默认值 | 发现后 24h |
| 响应格式重构 | 新增 DTO 版本适配器 | 发现后 48h |
| 接口 URL 变更 | 更新 client 端配置 | 发现后 24h |

### 2.2 适配器模式

当数据源接口发生不兼容变更时，使用适配器隔离影响：

```go
// 旧解析器
type cwlParserV1 struct {}
func (p *cwlParserV1) Parse(data []byte) ([]model.Draw, error) { ... }

// 新解析器
type cwlParserV2 struct {}
func (p *cwlParserV2) Parse(data []byte) ([]model.Draw, error) { ... }

// 根据响应特征选择解析器
func selectParser(data []byte) parser {
    if containsField(data, "frontWinningNum") {
        return &cwlParserV1{}
    }
    return &cwlParserV2{}
}
```

## 3. 内部 API 规范（预留）

### 3.1 适用场景

当后续扩展以下功能时，需定义内部 API 版本规范：
- 自建数据分析服务端
- 用户推荐记录云端同步
- 多数据源聚合查询

### 3.2 版本化策略

| 版本标识方式 | 示例 |
|------------|------|
| URL 路径版本 | `/api/v1/draws` |
| Accept Header | `Accept: application/vnd.lottery.v1+json` |
| Query 参数 | `/api/draws?version=1` |

优先使用 URL 路径版本方式。

### 3.3 兼容性承诺

| 版本状态 | 说明 | 兼容要求 |
|---------|------|---------|
| Alpha | 内部开发版本，不对外 | 不承诺兼容 |
| Stable | 正式发布版本 | 大版本内向后兼容 |
| Deprecated | 标记废弃 | 保留 ≥2 个小版本周期 |
| Sunset | 完全下线 | 提前一个版本周期通知 |

## 4. 限流与重试策略

### 4.1 限流规则

```go
// client/cwl.go 内置限流
type CWLClient struct {
    client  *http.Client
    limiter *rate.Limiter  // 每秒最多 1 个请求
}

func NewCWLClient() *CWLClient {
    return &CWLClient{
        client: &http.Client{Timeout: 10 * time.Second},
        limiter: rate.NewLimiter(rate.Every(1*time.Second), 1),
    }
}
```

### 4.2 重试策略

| 条件 | 行为 |
|------|------|
| HTTP 5xx | 重试最多 3 次，间隔 1s/2s/4s（指数退避） |
| HTTP 429（限流） | 等待 `Retry-After` 头指定的时间后重试 |
| HTTP 4xx（非 429） | 不重试，直接返回错误 |
| 网络超时 | 重试最多 2 次 |
| context canceled | 不重试 |
