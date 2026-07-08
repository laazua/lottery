# 06-testing-validation.md — 测试与验证循环

## 1. 测试分层策略

采用经典测试金字塔模型：

```
        ╱╲
       ╱  ╲          UI / 集成测试 — 覆盖面广，执行慢
      ╱    ╲
     ╱──────╲
    ╱        ╲        Service 层测试 — 核心业务逻辑
   ╱          ╲
  ╱──────────────╲
 ╱                ╲    Client 层 + Model 层测试 — 快速、稳定
╱──────────────────╲
```

### 1.1 Client 层测试

| 测试类型 | 工具 | 目标覆盖率 |
|---------|------|-----------|
| HTTP mock 测试 | `httptest.Server` | 90%+ |
| JSON 解析测试 | 原生测试 | 100% |
| 限流/重试测试 | 模拟场景 | 核心路径 |

```go
// client/cwl_test.go
func TestCWLClient_FetchDraws(t *testing.T) {
    // 使用 httptest.NewServer mock 远程 API
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "GET", r.Method)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(mockResponse))
    }))
    defer server.Close()

    client := NewCWLClient(WithBaseURL(server.URL))
    draws, err := client.FetchDraws(context.Background())
    assert.NoError(t, err)
    assert.Len(t, draws, 20)
}
```

### 1.2 Service 层测试

| 测试类型 | 工具 | 目标覆盖率 |
|---------|------|-----------|
| 业务逻辑测试 | mock client | 85%+ |
| 错误传递测试 | mock 返回错误 | 涵盖所有错误码 |
| 边界条件测试 | 空数据、极值数据 | 按需 |

```go
// service/lottery_test.go
func TestLotteryService_CalculateStats(t *testing.T) {
    mockAPI := &mock.MockLotteryAPI{
        FetchDrawsFunc: func(ctx context.Context, opts ...client.Option) ([]model.Draw, error) {
            return testDraws, nil
        },
    }
    svc := NewLotteryService(mockAPI)
    stats, err := svc.CalculateStats(context.Background(), 50)
    assert.NoError(t, err)
    assert.Equal(t, 35, len(stats.FrontFrequencies))
    assert.Equal(t, 12, len(stats.BackFrequencies))
}
```

### 1.3 UI 层测试

| 测试类型 | 方法 | 覆盖策略 |
|---------|------|---------|
| 组件渲染测试 | gioui test helpers | 核心组件 |
| 状态转换测试 | 手动触发+断言 | 所有用户交互路径 |
| 错误展示测试 | mock service 返回错误 | 每种错误类型 |

> 注：gioui 的 UI 测试能力有限，重点关注 `service/` 和 `client/` 覆盖。

## 2. 测试覆盖率目标与度量

### 2.1 量化目标

| 层级 | 目标覆盖率 | 强制阻断 |
|------|-----------|---------|
| 全局 | ≥80% | CI 强制 |
| `client/` | ≥90% | CI 建议 |
| `service/` | ≥85% | CI 强制 |
| `model/` | 100% | code review |
| `ui/` | — | 手动验证 |

### 2.2 覆盖率报告

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看各包覆盖率
go tool cover -func=coverage.out

# HTML 查看
go tool cover -html=coverage.out -o coverage.html
```

### 2.3 覆盖率排除

以下代码不计入覆盖率统计：
- `main.go` 入口函数
- gioui 自动生成的代码
- 纯 getter/setter（注解排除）

## 3. Mock 策略

### 3.1 Mock 接口定义

所有需要 mock 的接口定义在 `client/` 包中：

```go
// client/mock/mock_lottery.go

package mock

type MockLotteryAPI struct {
    FetchDrawsFunc        func(ctx context.Context, opts ...client.Option) ([]model.Draw, error)
    FetchLatestDrawFunc   func(ctx context.Context) (*model.Draw, error)
}
```

### 3.2 Mock 使用模式

```go
func TestStatsCalculation(t *testing.T) {
    mock := &mock.MockLotteryAPI{
        FetchDrawsFunc: func(ctx context.Context, opts ...client.Option) ([]model.Draw, error) {
            return generateTestDraws(100), nil
        },
    }
    
    statsSvc := service.NewStatsService(mock)
    result, err := statsSvc.CalculateStats(context.Background(), 50)
    
    assert.NoError(t, err)
    assert.Greater(t, result.FrontFrequencies[0].Count, 0)
}
```

## 4. CI 验证循环

### 4.1 提交前本地验证

```bash
# Step 1: 格式化检查
gofmt -l -d .
# 或
goimports -l .

# Step 2: 静态分析
go vet ./...

# Step 3: 测试 + 覆盖率
go test -count=1 -race -coverprofile=coverage.out ./...

# Step 4: 检查覆盖率阈值
go tool cover -func=coverage.out | grep total | awk '{print $3}' | cut -d. -f1
# 需 ≥ 80
```

### 4.2 CI 流水线

```
Push / PR 触发
    │
    ▼
┌──────────────────┐
│ gofmt 格式检查    │─── 阻断 ❌ → 开发者修正
└──────┬───────────┘
       ▼ 通过
┌──────────────────┐
│ go vet 静态分析   │─── 阻断 ❌ → 开发者修正
└──────┬───────────┘
       ▼ 通过
┌──────────────────┐
│ go test 单元测试  │─── 阻断 ❌ → 开发者修正
└──────┬───────────┘
       ▼ 通过
┌──────────────────┐
│ 覆盖率 ≥ 80% 检查  │─── 阻断 ❌ → 补充测试
└──────┬───────────┘
       ▼ 通过
┌──────────────────┐
│ go build 构建验证  │─── 阻断 ❌ → 开发者修正
└──────┬───────────┘
       ▼ 通过
    ✅ 验证通过
```

### 4.3 CI 脚本示例

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26.1'

      - name: Format check
        run: test -z $(gofmt -l .)

      - name: Vet
        run: go vet ./...

      - name: Test with coverage
        run: go test -count=1 -race -coverprofile=coverage.out ./...

      - name: Check coverage threshold
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
          if [ "$coverage" -lt 80 ]; then
            echo "Coverage $coverage% < 80%, failing"
            exit 1
          fi

      - name: Build
        run: go build ./...
```

## 5. 测试示例模板

### 5.1 Service 层测试模板

```go
package service_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/user/lottery/client/mock"
    "github.com/user/lottery/service"
)

func TestXxxService_MethodName(t *testing.T) {
    // Arrange
    mockClient := &mock.MockLotteryAPI{
        FetchDrawsFunc: func(ctx context.Context, opts ...client.Option) ([]model.Draw, error) {
            return testdata.SomeDraws, nil
        },
    }
    svc := service.NewXxxService(mockClient)

    // Act
    result, err := svc.SomeMethod(context.Background(), arg1)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, expected, result.SomeField)
}
```

### 5.2 Client 层测试模板

```go
package client_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/user/lottery/client"
)

func TestCWLClient_FetchXxx(t *testing.T) {
    // Arrange: mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 验证请求参数
        assert.Equal(t, "GET", r.Method)
        // 返回 mock 数据
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"total":1,"data":[...]}`))
    }))
    defer server.Close()

    c := client.NewCWLClient(client.WithBaseURL(server.URL))

    // Act
    result, err := c.FetchXxx(context.Background())

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, result)
}
```

## 6. 测试数据管理

- 测试数据统一存放在 `internal/testdata/` 目录
- JSON mock 响应存放在 `client/testdata/` 下
- 大体积测试数据用 `.json` 文件管理，不内嵌在代码中
