// Package config 提供构建时配置变量，通过 ldflags 注入。
//
// 使用方式：
//
//	go run -ldflags="-X 'github.com/user/lottery/internal/config.APIBaseURL=https://api.example.com'" .
//
//	go run -ldflags="-X 'github.com/user/lottery/internal/config.DataSource=mock'" .
package config

var (
	// APIBaseURL 数据源 API 基础 URL。
	// 通过 -ldflags 在构建时覆盖：
	//   -X 'github.com/user/lottery/internal/config.APIBaseURL=https://api.example.com'
	APIBaseURL = "https://www.cwl.gov.cn"

	// DataSource 数据源类型。
	// "cwl"  = 福彩官网接口（实时数据）
	// "mock" = 内置模拟数据（离线可用，无网络依赖）
	DataSource = "cwl"
)
