package model

import (
	"testing"
)

func TestNumberFrequency_Fields(t *testing.T) {
	nf := NumberFrequency{
		Number:    1,
		Count:     10,
		Frequency: 0.25,
		MissValue: 3,
		MaxMiss:   8,
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"号码", nf.Number, 1},
		{"出现次数", nf.Count, 10},
		{"频率", nf.Frequency, 0.25},
		{"遗漏值", nf.MissValue, 3},
		{"最大遗漏", nf.MaxMiss, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestStatistics_Fields(t *testing.T) {
	stats := Statistics{
		FrontHot:  []NumberFrequency{{Number: 1, Count: 10}},
		FrontWarm: []NumberFrequency{{Number: 2, Count: 5}},
		FrontCold: []NumberFrequency{{Number: 3, Count: 2}},
		BackHot:   []NumberFrequency{{Number: 6, Count: 8}},
		BackWarm:  []NumberFrequency{{Number: 7, Count: 4}},
		BackCold:  []NumberFrequency{{Number: 8, Count: 1}},
	}

	tests := []struct {
		name string
		got  int
		want int
	}{
		{"前区热号数量", len(stats.FrontHot), 1},
		{"前区温号数量", len(stats.FrontWarm), 1},
		{"前区冷号数量", len(stats.FrontCold), 1},
		{"后区热号数量", len(stats.BackHot), 1},
		{"后区温号数量", len(stats.BackWarm), 1},
		{"后区冷号数量", len(stats.BackCold), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestFrequencyResult_Fields(t *testing.T) {
	fr := FrequencyResult{
		FrontFrequencies: []NumberFrequency{{Number: 1, Count: 10}},
		BackFrequencies:  []NumberFrequency{{Number: 6, Count: 8}},
	}

	if len(fr.FrontFrequencies) != 1 {
		t.Errorf("expected 1 front frequency, got %d", len(fr.FrontFrequencies))
	}
	if len(fr.BackFrequencies) != 1 {
		t.Errorf("expected 1 back frequency, got %d", len(fr.BackFrequencies))
	}
}
