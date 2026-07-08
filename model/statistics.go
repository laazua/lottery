package model

// NumberFrequency 记录单个号码的频次统计。
type NumberFrequency struct {
	Number    int     // 号码
	Count     int     // 出现次数
	Frequency float64 // 出现频率（占比）
	MissValue int     // 当前遗漏值
	MaxMiss   int     // 历史最大遗漏值
}

// FrequencyResult 包含前区和后区的频次统计结果。
type FrequencyResult struct {
	FrontFrequencies []NumberFrequency // 前区 1-35 频次
	BackFrequencies  []NumberFrequency // 后区 1-12 频次
}

// Statistics 包含完整的冷热号统计结果。
type Statistics struct {
	FrontHot  []NumberFrequency // 前区热号
	FrontWarm []NumberFrequency // 前区温号
	FrontCold []NumberFrequency // 前区冷号
	BackHot   []NumberFrequency // 后区热号
	BackWarm  []NumberFrequency // 后区温号
	BackCold  []NumberFrequency // 后区冷号
}
