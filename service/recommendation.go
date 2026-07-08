package service

import (
	"context"
	"log/slog"
	"math/rand"
	"sort"

	"github.com/user/lottery/client"
	"github.com/user/lottery/internal/errors"
	"github.com/user/lottery/model"
)

// RecommendService 提供智能号码推荐。
type RecommendService struct {
	api client.LotteryAPI
}

// NewRecommendService 创建 RecommendService 实例。
func NewRecommendService(api client.LotteryAPI) *RecommendService {
	return &RecommendService{api: api}
}

// GenerateRecommendation 基于冷热统计和遗漏值生成推荐号码。
func (s *RecommendService) GenerateRecommendation(ctx context.Context, stats *model.Statistics) (*model.Recommendation, error) {
	if stats == nil {
		return nil, errors.ErrNoValidRecommendation
	}

	rec := &model.Recommendation{
		FrontNumbers: make([]model.RecommendNumber, 0, 5),
		BackNumbers:  make([]model.RecommendNumber, 0, 2),
	}

	selected := make(map[int]bool)

	hotPicks := pickWeighted(stats.FrontHot, 3, selected)
	for _, p := range hotPicks {
		selected[p.Number] = true
		rec.FrontNumbers = append(rec.FrontNumbers, model.RecommendNumber{
			Number: p.Number,
			Reason: "🔥热",
		})
	}

	warmPicks := pickWeighted(stats.FrontWarm, 1, selected)
	for _, p := range warmPicks {
		selected[p.Number] = true
		rec.FrontNumbers = append(rec.FrontNumbers, model.RecommendNumber{
			Number: p.Number,
			Reason: "🌡️温",
		})
	}

	coldPicks := pickMissed(stats.FrontCold, 1, selected)
	for _, p := range coldPicks {
		selected[p.Number] = true
		rec.FrontNumbers = append(rec.FrontNumbers, model.RecommendNumber{
			Number: p.Number,
			Reason: "📊遗漏",
		})
	}

	backPool := append(stats.BackHot, stats.BackWarm...)
	backPicks := pickWeighted(backPool, 1, selected)
	for _, p := range backPicks {
		selected[p.Number] = true
		rec.BackNumbers = append(rec.BackNumbers, model.RecommendNumber{
			Number: p.Number,
			Reason: "🔥热",
		})
	}

	backColdPicks := pickMissed(stats.BackCold, 1, selected)
	for _, p := range backColdPicks {
		rec.BackNumbers = append(rec.BackNumbers, model.RecommendNumber{
			Number: p.Number,
			Reason: "📊遗漏",
		})
	}

	slog.Info("推荐生成完成",
		"frontCount", len(rec.FrontNumbers),
		"backCount", len(rec.BackNumbers),
	)

	return rec, nil
}

// pickWeighted 从号码列表中加权随机选择 count 个（去重）。
func pickWeighted(freqs []model.NumberFrequency, count int, excluded map[int]bool) []model.NumberFrequency {
	if len(freqs) == 0 || count <= 0 {
		return nil
	}

	var totalWeight int
	for _, f := range freqs {
		if !excluded[f.Number] {
			totalWeight += f.Count + 1
		}
	}

	if totalWeight == 0 {
		return nil
	}

	result := make([]model.NumberFrequency, 0, count)
	for len(result) < count && len(result) < len(freqs) {
		r := rand.Intn(totalWeight)
		cumulative := 0
		for _, f := range freqs {
			if excluded[f.Number] {
				continue
			}
			cumulative += f.Count + 1
			if r < cumulative {
				result = append(result, f)
				excluded[f.Number] = true
				totalWeight -= f.Count + 1
				break
			}
		}
	}

	return result
}

// pickMissed 从号码列表中选择遗漏值最高的 count 个（去重）。
func pickMissed(freqs []model.NumberFrequency, count int, excluded map[int]bool) []model.NumberFrequency {
	if len(freqs) == 0 || count <= 0 {
		return nil
	}

	sorted := make([]model.NumberFrequency, len(freqs))
	copy(sorted, freqs)
	sort.Slice(sorted, func(i, j int) bool {
		if excluded[sorted[i].Number] && !excluded[sorted[j].Number] {
			return false
		}
		if !excluded[sorted[i].Number] && excluded[sorted[j].Number] {
			return true
		}
		return sorted[i].MissValue > sorted[j].MissValue
	})

	result := make([]model.NumberFrequency, 0, count)
	for _, f := range sorted {
		if excluded[f.Number] {
			continue
		}
		if len(result) >= count {
			break
		}
		result = append(result, f)
	}

	return result
}
