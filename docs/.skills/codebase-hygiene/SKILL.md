# SKILL.md — 代码库卫生清理

## 元信息

- **名称**: codebase-hygiene
- **用途**: 定期执行代码库质量检查，确保项目符合 AGENTS.md 中定义的工程规范
- **触发场景**:
  - 每两周例行健康检查
  - 代码提交前发现代码异味
  - 新增外部依赖后
  - CI 流水线失败时定位问题

## 前置条件

- [ ] 当前在项目根目录 `/opt/codes/lottery`
- [ ] 工作区无未提交的修改（或已确认不冲突）
- [ ] 已安装 Go 1.26.1 及所需工具链
- [ ] 依赖已安装：`go mod tidy` 已执行

---

## 执行步骤

### 1. Go 源码格式检查

**操作**：

```bash
# 检查未格式化的文件
gofmt -l .

# 如有输出文件列表，格式化它们
gofmt -w <file1> <file2>

# 或一次性修复全部
gofmt -w .

# 使用 goimports（推荐，gofmt 超集）
goimports -l .
```

**验证**：再次运行 `gofmt -l .` 应无输出。

---

### 2. 静态分析检查

**操作**：

```bash
# 运行 go vet
go vet ./...

# 检查所有包
go vet ./client/... ./service/... ./ui/... ./model/... ./internal/...
```

**验证**：`go vet` 应无输出（无错误）。

**常见问题**：
- 循环引用 → 检查 `docs/03-boundary-constraints.md` 依赖方向
- 未使用的变量 → 删除或使用 `_`

---

### 3. 依赖健康检查

**操作**：

```bash
# 清理依赖
go mod tidy

# 检查依赖是否有漏洞
go mod verify
```

**验证**：
- `go mod tidy` 不报错
- `go mod verify` 输出应为空或 `all modules verified`

**外部依赖登记检查**：
- [ ] 检查是否有新增的外部包未在 `docs/03-boundary-constraints.md` 中登记
- [ ] 如果是，补充到"已登记的外部依赖"表中

```bash
# 列出所有直接依赖
grep -v indirect go.mod | grep -E '^\t' | awk '{print $1, $2}'
# 与 docs/03-boundary-constraints.md 的登记表比对
```

---

### 4. 错误处理检查

**操作**：检查是否有使用 `_` 丢弃 error 的情况。

```bash
# 搜索 _ 丢弃 error 的模式
grep -rn '_ := ' --include='*.go' --exclude='*_test.go' .

# 搜索 fmt.Println 使用
grep -rn 'fmt\.Print' --include='*.go' .
```

**验证**：
- [ ] 无 `_ :=` 丢弃 error 的代码（测试文件除外）
- [ ] 无 `fmt.Println` / `fmt.Printf` 使用（应用代码中）
- [ ] 所有错误处理使用 slog + error wrapping

---

### 5. 文档注释检查

**操作**：

```bash
# 列出所有没有文档注释的导出名称
go list ./... | xargs -n1 go doc -all 2>/dev/null | grep -B1 '^func \|^type \|^var \|^const ' | grep -v '^--$' | head -50
```

**手动检查示例**：

```go
// ✅ 正确
// FetchDraws 从公开 API 拉取指定范围的开奖数据。
func FetchDraws(...)

// ❌ 错误
// 拉取数据
func FetchDraws(...)

// ❌ 缺失
func FetchDraws(...)
```

**验证**：
- [ ] 每个导出的（大写开头）类型、函数、方法、变量都有文档注释
- [ ] 注释以名称开头、句号结尾

---

### 6. 命名规范检查

**操作**：

```bash
# 检查包名与目录名的一致性
# 列出所有包声明和目录路径
find . -name '*.go' -not -path '*/vendor/*' | grep -v '_test.go' | while read f; do
    dir=$(dirname "$f")
    pkg=$(head -1 "$f" | grep '^package ')
    if [ -n "$pkg" ]; then
        echo "$dir → $pkg"
    fi
done | sort -u
```

**检查项**：
- [ ] 包名与目录名一致
- [ ] 包名为单数小写
- [ ] 文件名使用小写字母和下划线
- [ ] 测试文件以 `_test.go` 结尾

---

### 7. 测试与覆盖率检查

**操作**：

```bash
# 运行全部测试
go test -count=1 -race ./...

# 生成覆盖率报告
go test -count=1 -coverprofile=coverage.out ./...

# 查看各包覆盖率
go tool cover -func=coverage.out

# 提取总覆盖率
go tool cover -func=coverage.out | grep total | awk '{print $3}'
```

**验证**：
- [ ] 全部测试通过（`FAIL` 计数为 0）
- [ ] 总覆盖率 ≥ 80%
- [ ] `client/` 包覆盖率 ≥ 90%
- [ ] `service/` 包覆盖率 ≥ 85%
- [ ] Race 条件检测通过（`-race` 不报错）

---

### 8. 构建验证

**操作**：

```bash
# 构建项目
go build ./...
```

**验证**：构建成功，无编译错误。

---

### 9. client 层封装检查

**操作**：

```go
// 检查所有外部 HTTP 请求是否经过 client/ 层
grep -rn '"http\.\|net/http' --include='*.go' . | grep -v '_test.go' | grep -v 'client/' | grep -v 'internal/httputil'
```

**验证**：
- [ ] 所有 `net/http` 的 import 仅在 `client/` 和 `internal/httputil` 中出现
- [ ] `service/` 和 `ui/` 不直接发起 HTTP 请求
- [ ] 所有外部 API 依赖通过接口抽象

---

### 10. 敏感信息检查

**操作**：

```bash
# 检查是否意外提交了密钥或凭据
grep -rn 'api[_-]key\|apikey\|secret\|token\|password\|passwd' --include='*.go' --include='*.env' --include='*.yaml' --include='*.json' .

# 检查 .gitignore 是否覆盖了敏感文件
cat .gitignore 2>/dev/null || echo "⚠️ .gitignore 不存在"
```

**验证**：
- [ ] 无 API 密钥、Token、密码硬编码在代码中
- [ ] `.gitignore` 存在且覆盖了 `*.env`、`*secret*` 等模式

---

## 修复策略

### 格式问题

```bash
# 自动修复全部格式
gofmt -w .
# 或
goimports -w .
```

### 依赖问题

```bash
# 清理无效依赖
go mod tidy
```

### 测试失败

1. 定位失败测试：`go test -v ./...` 查看具体哪个断言失败
2. 如果是 mock 不匹配：检查 mock 返回值与实际调用是否一致
3. 如果是覆盖不足：补充对应的测试用例

### 文档注释缺失

逐个文件检查 `go doc` 输出，为缺失注释的导出名称补充文档注释。

---

## 验证确认

清理完成后，执行最终验证：

```bash
echo "=== 1. 格式检查 ===" && gofmt -l . && \
echo "=== 2. 静态分析 ===" && go vet ./... && \
echo "=== 3. 测试 ===" && go test -count=1 -race -coverprofile=coverage.out ./... && \
echo "=== 4. 覆盖率 ===" && go tool cover -func=coverage.out | grep total && \
echo "=== 5. 构建 ===" && go build ./... && \
echo "=== 6. 依赖 ===" && go mod verify && \
echo "✅ 卫生清理完成"
```

---

## CI 集成

本 Skill 中的步骤已部分集成到 CI 流水线中（详见 `docs/06-testing-validation.md` CI 章节）。所有阻断项在 CI 中自动执行。

### CI 与 Skill 的边界

| 检查项 | CI 自动 | Skill 手动 | 说明 |
|--------|---------|-----------|------|
| gofmt | ✅ | — | CI 强制阻断 |
| go vet | ✅ | — | CI 强制阻断 |
| go test | ✅ | — | CI 强制阻断 |
| 覆盖率 | ✅ | — | CI 强制阻断 |
| go build | ✅ | — | CI 强制阻断 |
| 错误处理检查 | — | ✅ | 人工 code review |
| 命名规范 | — | ✅ | 人工 code review |
| client 封装检查 | — | ✅ | 人工 code review |
| 敏感信息检查 | — | ✅ | 人工 code review |
| 文档同步检查 | — | ✅ | 人工确认 |
| 外部依赖登记 | — | ✅ | 人工确认 |

---

## 常见问题速查

| 症状 | 可能原因 | 修复指引 |
|------|---------|---------|
| `gofmt -l .` 有输出 | 文件未格式化 | 执行 `gofmt -w .` |
| `go vet` 报循环引用 | 包依赖方向违规 | 查看 `docs/03-boundary-constraints.md` |
| `go test` 有 FAIL | 代码变动导致测试断裂 | 查看具体失败断言 |
| 覆盖率 < 80% | 新代码缺少测试 | 为 service/ 和 client/ 新增测试 |
| `go build` 编译错误 | 语法错误或缺失依赖 | 检查报错行号修复 |
