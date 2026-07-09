// Package config 提供构建时配置变量，通过 ldflags 注入。
//
// 使用方式：
//
//	go run -ldflags="-X 'github.com/user/lottery/internal/config.APIBaseURL=https://api.example.com'" .
package config

var (
	// APIBaseURL 数据源 API 基础 URL。
	// 通过 -ldflags 在构建时覆盖：
	//   -X 'github.com/user/lottery/internal/config.APIBaseURL=https://api.example.com'
	APIBaseURL = "https://webapi.sporttery.cn"
)
