package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/user/lottery/client"
	"github.com/user/lottery/internal/errors"
	"github.com/user/lottery/model"
)

// StatsService 提供冷热号统计分析。
type StatsService struct {
	api client.LotteryAPI
}

// NewStatsService 创建 StatsService 实例。
func NewStatsService(api client.LotteryAPI) *StatsService {
	return &StatsService{api: api}
}

// CalculateStats 对指定期数范围的开奖数据进行冷热号统计。
// periodCount 指定统计多少期数据（20/50/100 期）。
func (s *StatsService) CalculateStats(ctx context.Context, draws []model.Draw, periodCount int) (*model.Statistics, error) {
	if len(draws) == 0 {
		return nil, fmt.Errorf("%w: 无开奖数据", errors.ErrInsufficientDraws)
	}
	if periodCount <= 0 {
		return nil, fmt.Errorf("%w: 期数必须大于 0", errors.ErrInvalidStatsRange)
	}

	targetDraws := draws
	if len(draws) > periodCount {
		targetDraws = draws[:periodCount]
	}

	frontFreq := make(map[int]int)
	backFreq := make(map[int]int)
	lastFrontSeen := make(map[int]int)
	lastBackSeen := make(map[int]int)

	for i, draw := range targetDraws {
		for _, n := range draw.FrontNumbers {
			frontFreq[n]++
			lastFrontSeen[n] = i
		}
		for _, n := range draw.BackNumbers {
			backFreq[n]++
			lastBackSeen[n] = i
		}
	}

	totalDraws := len(targetDraws)

	frontResult := buildFrequencies(1, 35, frontFreq, lastFrontSeen, totalDraws)
	backResult := buildFrequencies(1, 12, backFreq, lastBackSeen, totalDraws)

	sortFrequencies(frontResult)
	sortFrequencies(backResult)

	stats := &model.Statistics{}
	stats.FrontHot, stats.FrontWarm, stats.FrontCold = splitHotCold(frontResult)
	stats.BackHot, stats.BackWarm, stats.BackCold = splitHotCold(backResult)

	slog.Info("冷热统计完成",
		"draws", totalDraws,
		"frontHot", len(stats.FrontHot),
		"frontWarm", len(stats.FrontWarm),
		"frontCold", len(stats.FrontCold),
	)

	return stats, nil
}

// buildFrequencies 为指定号码范围构建频次统计结果。
func buildFrequencies(min, max int, freqMap map[int]int, lastSeen map[int]int, totalDraws int) []model.NumberFrequency {
	result := make([]model.NumberFrequency, 0, max-min+1)
	for n := min; n <= max; n++ {
		count := freqMap[n]
		var freq float64
		if totalDraws > 0 {
			freq = float64(count) / float64(totalDraws)
		}
		missValue := totalDraws - lastSeen[n] - 1
		if missValue < 0 {
			missValue = 0
		}
		result = append(result, model.NumberFrequency{
			Number:    n,
			Count:     count,
			Frequency: freq,
			MissValue: missValue,
			MaxMiss:   missValue,
		})
	}
	return result
}

// sortFrequencies 按频次降序排列。
func sortFrequencies(freqs []model.NumberFrequency) {
	sort.Slice(freqs, func(i, j int) bool {
		return freqs[i].Count > freqs[j].Count
	})
}

// splitHotCold 将频次结果分为冷温热三档。
// 前 30% → 热号，后 30% → 冷号，中间 → 温号。
func splitHotCold(freqs []model.NumberFrequency) (hot, warm, cold []model.NumberFrequency) {
	n := len(freqs)
	if n == 0 {
		return nil, nil, nil
	}

	hotCount := n * 30 / 100
	if hotCount < 1 {
		hotCount = 1
	}
	coldCount := n * 30 / 100
	if coldCount < 1 {
		coldCount = 1
	}
	warmCount := n - hotCount - coldCount

	hot = freqs[:hotCount]
	warm = freqs[hotCount : hotCount+warmCount]
	cold = freqs[hotCount+warmCount:]

	return hot, warm, cold
}
