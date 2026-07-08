package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestError_Error(t *testing.T) {
	e := NewError("E0GN001", "未知错误")
	want := "[E0GN001] 未知错误"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_Is_SameCode(t *testing.T) {
	e1 := ErrNetworkUnreachable
	e2 := NewError("E1NT001", "网络不可达")
	if !errors.Is(e1, e2) {
		t.Errorf("errors.Is(e1, e2) should be true for same code")
	}
}

func TestError_Is_DifferentCode(t *testing.T) {
	if errors.Is(ErrUnknown, ErrInvalidParams) {
		t.Errorf("ErrUnknown should not match ErrInvalidParams")
	}
}

func TestError_Is_Wrapped(t *testing.T) {
	base := ErrNetworkUnreachable
	wrapped := fmt.Errorf("请求失败: %w", base)
	if !errors.Is(wrapped, ErrNetworkUnreachable) {
		t.Errorf("wrapped error should match ErrNetworkUnreachable")
	}
}

func TestNewError_AllCodes(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		code string
	}{
		{"ErrUnknown", ErrUnknown, "E0GN001"},
		{"ErrInvalidParams", ErrInvalidParams, "E0GN002"},
		{"ErrUnsupported", ErrUnsupported, "E0GN003"},
		{"ErrNetworkUnreachable", ErrNetworkUnreachable, "E1NT001"},
		{"ErrRequestTimeout", ErrRequestTimeout, "E1NT002"},
		{"ErrRateLimited", ErrRateLimited, "E1NT003"},
		{"ErrServerError", ErrServerError, "E1NT004"},
		{"ErrTooManyRetries", ErrTooManyRetries, "E1NT005"},
		{"ErrParseResponse", ErrParseResponse, "E2DP001"},
		{"ErrUnexpectedField", ErrUnexpectedField, "E2DP002"},
		{"ErrEmptyResponse", ErrEmptyResponse, "E2DP003"},
		{"ErrServiceUnavailable", ErrServiceUnavailable, "E2SV001"},
		{"ErrInsufficientDraws", ErrInsufficientDraws, "E3ST001"},
		{"ErrInvalidStatsRange", ErrInvalidStatsRange, "E3ST002"},
		{"ErrNoValidRecommendation", ErrNoValidRecommendation, "E3RC001"},
		{"ErrRecommendationDisabled", ErrRecommendationDisabled, "E3RC002"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("code = %q, want %q", tt.err.Code, tt.code)
			}
			if tt.err.Message == "" {
				t.Errorf("message should not be empty")
			}
		})
	}
}
