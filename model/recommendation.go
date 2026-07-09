package model

// RecommendNumber 表示单个推荐号码及其理由。
type RecommendNumber struct {
	Number int    // 号码
	Reason string // 推荐理由标签：热号 / 温号 / 遗漏
}

// Recommendation 包含一组推荐号码组合。
type Recommendation struct {
	FrontNumbers []RecommendNumber // 前区推荐号码（5个）
	BackNumbers  []RecommendNumber // 后区推荐号码（2个）
}
