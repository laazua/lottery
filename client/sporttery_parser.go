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

// sportteryResponse 是 webapi.sporttery.cn 响应顶层结构。
type sportteryResponse struct {
	Value   sportteryValue `json:"value"`
	Total   int            `json:"total"`
	Success bool           `json:"success"`
}

// sportteryValue 包含开奖数据列表及分页信息。
type sportteryValue struct {
	List  []sportteryDrawDTO `json:"list"`
	Total int                `json:"total"` // 总期数
}

// sportteryDrawDTO 是单期开奖数据的 DTO。
// 字段名映射自体彩官方 API 响应。
type sportteryDrawDTO struct {
	LotteryDrawNum          string `json:"lotteryDrawNum"`          // 期号
	LotteryDrawTime         string `json:"lotteryDrawTime"`         // 开奖日期
	LotterySaleEndtime      string `json:"lotterySaleEndtime"`      // 销售截止时间
	LotteryDrawResult       string `json:"lotteryDrawResult"`       // 开奖号码（空格分隔：前5后2）
	LotteryUnsortDrawresult string `json:"lotteryUnsortDrawresult"` // 未排序开奖号码
	TotalSaleAmount         string `json:"totalSaleAmount"`         // 销售额（含逗号，如"303,115,587"）
	PoolBalance             string `json:"poolBalance"`             // 奖池金额（含逗号，如"814,593,833.81"）
	LotteryEquipmentCount   int    `json:"lotteryEquipmentCount"`   // 摇奖机编号
}

// parseSportteryResponse 解析体彩 API JSON 响应为 Draw 列表及总记录数。
func parseSportteryResponse(data []byte) ([]model.Draw, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.ErrEmptyResponse
	}

	var resp sportteryResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, 0, fmt.Errorf("%w: %s", errors.ErrParseResponse, err.Error())
	}

	if !resp.Success {
		return nil, 0, fmt.Errorf("%w: API 返回失败状态", errors.ErrAPIResponse)
	}

	if len(resp.Value.List) == 0 {
		return []model.Draw{}, resp.Value.Total, nil
	}

	draws := make([]model.Draw, 0, len(resp.Value.List))
	for _, dto := range resp.Value.List {
		draw, err := convertSportteryDTO(dto)
		if err != nil {
			slog.Warn("跳过解析失败的开奖数据", "issue", dto.LotteryDrawNum, "error", err)
			continue
		}
		draws = append(draws, *draw)
	}

	return draws, resp.Value.Total, nil
}

// convertSportteryDTO 将体彩 DTO 转换为 model.Draw。
func convertSportteryDTO(dto sportteryDrawDTO) (*model.Draw, error) {
	// 解析开奖时间（优先使用 lotteryDrawTime，回退到 lotterySaleEndtime）
	var drawTime time.Time
	var err error
	drawTime, err = time.Parse("2006-01-02", dto.LotteryDrawTime)
	if err != nil {
		drawTime, err = time.Parse("2006-01-02 15:04:05", dto.LotterySaleEndtime)
		if err != nil {
			return nil, fmt.Errorf("解析开奖时间失败: %w", err)
		}
	}

	// 解析开奖号码：lotteryDrawResult 格式为 "05 12 18 23 31 07 09"
	// 前 5 个为前区号码，后 2 个为后区号码
	result := dto.LotteryDrawResult
	if result == "" {
		result = dto.LotteryUnsortDrawresult
	}
	if result == "" {
		return nil, fmt.Errorf("%w: 开奖号码为空", errors.ErrUnexpectedField)
	}

	parts := strings.Fields(result)
	if len(parts) != 7 {
		return nil, fmt.Errorf("%w: 号码数量不匹配，期望 7 个，实际 %d 个", errors.ErrUnexpectedField, len(parts))
	}

	frontNums, err := parseSportteryNumbers(parts[:5], 5)
	if err != nil {
		return nil, fmt.Errorf("解析前区号码: %w", err)
	}

	backNums, err := parseSportteryNumbers(parts[5:7], 2)
	if err != nil {
		return nil, fmt.Errorf("解析后区号码: %w", err)
	}

	return &model.Draw{
		Issue:        dto.LotteryDrawNum,
		DrawTime:     drawTime,
		FrontNumbers: [5]int{frontNums[0], frontNums[1], frontNums[2], frontNums[3], frontNums[4]},
		BackNumbers:  [2]int{backNums[0], backNums[1]},
		SaleAmount:   parseAmount(dto.TotalSaleAmount),
		PoolAmount:   parseAmount(dto.PoolBalance),
	}, nil
}

// parseSportteryNumbers 解析字符串切片为整数切片。
func parseSportteryNumbers(parts []string, expectedLen int) ([]int, error) {
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

// parseAmount 解析含逗号的金额字符串为 int64（单位：元）。
// 如 "303,115,587" → 303115587，"814,593,833.81" → 814593833（取整）。
func parseAmount(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0
	}
	// 可能含小数，按小数点取整
	if idx := strings.Index(s, "."); idx >= 0 {
		s = s[:idx]
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
