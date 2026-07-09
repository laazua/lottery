# AGENTS.md - 项目地图

手机移动端一款大乐透往期数据查询，分析，以及推荐选号的APP；主要技术栈Go + gioui + jdk

## 技术栈（不允许擅自修改）
- Go: version go1.26.1 linux/amd64
- gioui: v0.10.0
- java: 1.8.0_492

## 目录地图

| 想做什么 | 去哪里看 |
|----------|----------|
| 了解项目架构 | docs/01-architecture.md |
| 了解编码规范 | docs/02-coding-guidelines.md |
| 了解包边界与执行约束 | docs/03-boundary-constraints.md |
| 了解 API 版本规范 | docs/04-api-version-specification.md |
| 了解错误码规范 | docs/05-errors-code.md |
| 了解测试与验证循环 | docs/06-testing-validation.md |
| 了解 Harness 驾驭工程设计 | docs/07-harnesss.md |
| 移动端开发环境依赖环境 | docs/08-mobile-dev-deps.md |
| 执行代码库卫生清理 | docs/.skills/codebase-hygiene/SKILL.md |
| 了解开奖查询页面布局 | docs/11-history-page-layout.md |
| 了解开奖查询页面优化记录 | docs/12-history-page-optimization.md |


## 硬性规则
1. 代码必须通过 gofmt 格式化，建议使用 goimports 管理 import — lint 强制阻断；gofmt 是 Go 官方强制要求的格式化工具，几乎所有 Go 代码都使用它来统一风格。goimports 是 gofmt 的超集，可自动增删 import 行。CI 中应配置检查，确保提交代码均通过格式化。
2. 导出的包、类型、函数、方法必须有文档注释，注释以名称开头、完整句子结尾 — lint 强制阻断；所有顶层导出名称必须有文档注释，注释应为完整句子，以被注释的名称开头，以句号结尾。这确保 godoc 正确生成文档。
3. 错误处理必须显式检查，禁止用 _ 丢弃 error — lint 强制阻断；函数返回 error 时必须检查并处理，不能使用 _ 变量忽略错误。应返回 error 或在极端情况下使用 panic，但不要用 panic 处理普通错误。
4. 使用 log/slog 结构化日志，禁止使用 fmt.Println 等非结构化输出 — code review 强制；Go 1.21 提供了官方结构化日志包 log/slog，建议全项目统一使用。fmt.Println 不适合生产日志，无法被日志系统有效采集和查询，应通过 code review 禁止。
5. 包命名使用小写单数，与目录名一致；文件名小写，可用下划线分割，测试文件以 _test.go 结尾 — lint 建议 + code review包名应保持与目录一致，使用小写，不用下划线和复数。Go 文件名使用小写字母，测试文件必须以 _test.go 结尾。
6. 核心业务逻辑测试覆盖率 ≥ 80% — CI 强制阻断；现代 Go 项目通常要求测试覆盖率 ≥ 80%，这是许多开源项目和公司实践的常见标准。CI 应配置 go test -cover 检查覆盖率，不达标则阻断合并。
7. 外部依赖（数据库、第三方 API、中间件）必须通过 client/ 抽象层封装 — code review 强制；这是良好的工程实践，通过抽象层隔离外部依赖，便于单元测试 Mock 和依赖替换。具体实现模式（如 clients/ 目录）由团队约定，通过 code review 强制执行。