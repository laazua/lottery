package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/user/lottery/internal/errors"
	"github.com/user/lottery/model"
)

// cwlResponse 是 cwl.gov.cn API 响应的 DTO。
type cwlResponse struct {
	Total int          `json:"total"`
	Data  []cwlDrawDTO `json:"data"`
}

// cwlDrawDTO 是单期开奖数据 DTO。
type cwlDrawDTO struct {
	Issue           string `json:"issue"`
	DrawTime        string `json:"drawTime"`
	FrontWinningNum string `json:"frontWinningNum"`
	BackWinningNum  string `json:"backWinningNum"`
	SaleAmount      string `json:"saleAmount"`
	PoolAmount      string `json:"poolAmount"`
}

// parseDrawResponse 解析 cwl.gov.cn JSON 响应为 Draw 列表。
func parseDrawResponse(data []byte) ([]model.Draw, error) {
	if len(data) == 0 {
		return nil, errors.ErrEmptyResponse
	}

	var resp cwlResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrParseResponse, err.Error())
	}

	if len(resp.Data) == 0 {
		return []model.Draw{}, nil
	}

	draws := make([]model.Draw, 0, len(resp.Data))
	for _, dto := range resp.Data {
		draw, err := convertDTO(dto)
		if err != nil {
			slog.Warn("跳过解析失败的开奖数据", "issue", dto.Issue, "error", err)
			continue
		}
		draws = append(draws, *draw)
	}

	return draws, nil
}

// convertDTO 将单个 DTO 转换为 model.Draw。
func convertDTO(dto cwlDrawDTO) (*model.Draw, error) {
	drawTime, err := time.Parse("2006-01-02 15:04:05", dto.DrawTime)
	if err != nil {
		return nil, fmt.Errorf("解析开奖时间失败: %w", err)
	}

	frontNums, err := parseNumbers(dto.FrontWinningNum, 5)
	if err != nil {
		return nil, fmt.Errorf("解析前区号码: %w", err)
	}

	backParts, err := parseNumbers(dto.BackWinningNum, 2)
	if err != nil {
		return nil, fmt.Errorf("解析后区号码: %w", err)
	}

	saleAmount, _ := strconv.ParseInt(dto.SaleAmount, 10, 64)
	poolAmount, _ := strconv.ParseInt(dto.PoolAmount, 10, 64)

	return &model.Draw{
		Issue:        dto.Issue,
		DrawTime:     drawTime,
		FrontNumbers: [5]int{frontNums[0], frontNums[1], frontNums[2], frontNums[3], frontNums[4]},
		BackNumbers:  [2]int{backParts[0], backParts[1]},
		SaleAmount:   saleAmount,
		PoolAmount:   poolAmount,
	}, nil
}

// parseNumbers 解析逗号分隔的号码字符串为整数切片。
func parseNumbers(s string, expectedLen int) ([]int, error) {
	parts := strings.Split(s, ",")
	if len(parts) != expectedLen {
		return nil, fmt.Errorf("%w: 号码数量不匹配，期望 %d 个，实际 %d 个", errors.ErrUnexpectedField, expectedLen, len(parts))
	}

	nums := make([]int, expectedLen)
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("%w: 号码解析失败: %s", errors.ErrUnexpectedField, p)
		}
		nums[i] = n
	}

	return nums, nil
}
