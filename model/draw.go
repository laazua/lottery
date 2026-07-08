// Package model 定义大乐透核心数据模型，零外部依赖。
package model

import "time"

// Draw 表示单期大乐透开奖结果。
// FrontNumbers 为前区号码（1-35，升序排列，5个），
// BackNumbers 为后区号码（1-12，升序排列，2个）。
type Draw struct {
	Issue        string    // 期号，如 "24180"
	DrawTime     time.Time // 开奖日期
	FrontNumbers [5]int    // 前区号码
	BackNumbers  [2]int    // 后区号码
	SaleAmount   int64     // 销售额（元）
	PoolAmount   int64     // 奖池金额（元）
}
