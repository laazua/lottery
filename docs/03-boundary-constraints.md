# 03-boundary-constraints.md — 包边界与执行约束

## 1. 包依赖规则表

```
依赖方向: 左侧包 → 右侧包
符号说明: ✅ 允许  ❌ 禁止  ⚠️ 有条件（见注释）
```

| 源 \ 目标 | `model/` | `client/` | `service/` | `ui/` | `internal/` |
|------------|----------|-----------|------------|-------|-------------|
| `model/`   | —        | ❌        | ❌         | ❌    | ❌          |
| `client/`  | ✅       | —         | ❌         | ❌    | ✅          |
| `service/` | ✅       | ✅ (接口) | —          | ❌    | ✅          |
| `ui/`      | ✅       | ❌        | ✅         | —     | ❌          |
| `internal/`| ❌       | ❌        | ❌         | ❌    | —           |

### 注释说明

- **`service/ → client/（接口）允许，不依赖实现）**：`service` 层只能依赖 `client` 包中定义的接口（如 `LotteryAPI`），不能直接依赖 `client/cwl.go` 的具体实现。具体实现在 `main.go` 中通过 DI 注入。
- **`ui/ → model/ 允许**：仅限读取，不能修改 model 的定义。
- **`internal/` 不可被外部包导入**：Go 编译器对 `internal` 包有强制访问控制，项目根目录下的 `internal/` 仅能被该根目录下的包导入。

## 2. 禁止的依赖方向

### 2.1 硬性禁止

```go
// ❌ service 导入 ui
import "github.com/user/lottery/ui/screen"   // 禁止！

// ❌ model 导入任何项目包
import "github.com/user/lottery/service"       // 禁止！

// ❌ client 导入 service
import "github.com/user/lottery/service"       // 禁止！
```

### 2.2 循环引用检测

任何提交的代码不得引入循环依赖。在 `go vet` 中会自动检测：

```bash
go vet ./...
# 若存在循环引用，输出类似:
# import cycle not allowed in test
```

### 2.3 违规处理流程

1. CI 检测到循环引用 → 构建阻断
2. 开发者分析循环来源并重构
3. 常见修复：将共同依赖的类型抽取到 `model/` 包中

## 3. 接口隔离规则

### 3.1 面向接口编程

各层之间通过接口依赖，而非具体类型：

```go
// client/client.go — 定义接口
type LotteryAPI interface {
    FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error)
    FetchDrawByPeriod(ctx context.Context, period string) (*model.Draw, error)
}

// client/cwl.go — 具体实现
type CWLClient struct { ... }
func (c *CWLClient) FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error) { ... }
```

```go
// service/lottery.go — 依赖接口而非实现
type LotteryService struct {
    api client.LotteryAPI  // 依赖接口
}

func NewLotteryService(api client.LotteryAPI) *LotteryService {
    return &LotteryService{api: api}
}
```

### 3.2 接口最小化原则

接口应只包含调用方需要的方法，不贪多求全：

```go
// ✅ 恰到好处的接口
type DrawProvider interface {
    FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error)
}

// ❌ 过于臃肿的接口
type DrawProvider interface {
    FetchDraws(...)
    FetchDrawByPeriod(...)
    FetchLatestDraw(...)
    FetchUserInfo(...)   // 不需要
    FetchPrizePool(...)  // 不需要
}
```

## 4. Mock 策略

### 4.1 接口为 Mock 而生

所有 `client/` 层的接口设计时即需考虑可 Mock 性：

```go
// mock/client_mock.go — 测试用 Mock 实现
type MockLotteryAPI struct {
    FetchDrawsFunc func(ctx context.Context, opts ...Option) ([]model.Draw, error)
}

func (m *MockLotteryAPI) FetchDraws(ctx context.Context, opts ...Option) ([]model.Draw, error) {
    return m.FetchDrawsFunc(ctx, opts)
}
```

### 4.2 Mock 存放位置

- 测试用 Mock 放在与被测包同级的 `mock/` 子包中
- 例：`client/mock/mock.go`
- 或直接在被测包内部使用 `_test.go` 文件定义

## 5. 外部包引入审批

### 5.1 允许直接使用的标准库

- `fmt`, `errors`, `log/slog`, `encoding/json`
- `net/http`, `context`, `time`, `sync`
- `math`, `sort`, `strings`, `strconv`
- `io`, `os`, `path/filepath`

### 5.2 新外部包引入流程

1. 确认不可以标准库替代
2. 在 `docs/03-boundary-constraints.md` 中登记（未登记不审批）
3. 说明引入理由、版本、依赖范围
4. 团队讨论确认后，执行 `go get`

### 5.3 已登记的外部依赖

| 包 | 版本 | 用途 | 引入日期 | 类型 |
|----|------|------|---------|------|
| `gioui.org` | v0.10.0 | UI 框架 | — | 编译 |
| `gioui.org/x` | — | gioui 扩展组件 | — | 编译 |
| `github.com/stretchr/testify` | 最新 | 测试断言库 | — | 仅测试 |

## 6. 包引用规范检查清单

CI 和 code review 中使用以下清单检查合规性：

- [ ] 是否有违反依赖方向表的情况？
- [ ] `service/` 是否只通过接口依赖 `client/`？
- [ ] `model/` 是否零项目依赖？
- [ ] 是否有任何 `_` 丢弃了 error？
- [ ] 外部 API 调用是否全部通过 `client/` 层？
- [ ] 包名是否与目录名一致？
