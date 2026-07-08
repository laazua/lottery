// Package period 提供大乐透期号解析与验证工具。
package period

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/user/lottery/internal/errors"
)

// ValidatePeriod 验证期号格式。
// 期号格式为"年份后两位 + 序号"，如 "24180" 表示 2024 年第 180 期。
func ValidatePeriod(period string) error {
	if len(period) != 5 {
		return fmt.Errorf("%w: 期号长度必须为5位，当前 %d 位", errors.ErrInvalidParams, len(period))
	}
	if _, err := strconv.Atoi(period); err != nil {
		return fmt.Errorf("%w: 期号必须为数字: %s", errors.ErrInvalidParams, period)
	}
	return nil
}

// ExtractYear 从期号中提取年份（四位）。
func ExtractYear(period string) (int, error) {
	if err := ValidatePeriod(period); err != nil {
		return 0, err
	}
	prefix := period[:2]
	year, err := strconv.Atoi(prefix)
	if err != nil {
		return 0, fmt.Errorf("提取年份失败: %w", err)
	}
	return 2000 + year, nil
}

// ExtractSequence 从期号中提取当年序号。
func ExtractSequence(period string) (int, error) {
	if err := ValidatePeriod(period); err != nil {
		return 0, err
	}
	seq := strings.TrimLeft(period[2:], "0")
	if seq == "" {
		return 0, nil
	}
	return strconv.Atoi(seq)
}
