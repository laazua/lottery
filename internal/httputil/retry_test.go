package httputil

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNewHTTPClient_DefaultTimeout(t *testing.T) {
	client := NewHTTPClient(0)
	if client.Timeout != DefaultTimeout {
		t.Errorf("expected timeout %v, got %v", DefaultTimeout, client.Timeout)
	}
}

func TestNewHTTPClient_CustomTimeout(t *testing.T) {
	client := NewHTTPClient(5 * time.Second)
	if client.Timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", client.Timeout)
	}
}

func TestRetryWithBackoff_Success(t *testing.T) {
	attempts := 0
	err := RetryWithBackoff(context.Background(), 3, func() error {
		attempts++
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt on success, got %d", attempts)
	}
}

func TestRetryWithBackoff_RetryThenSuccess(t *testing.T) {
	attempts := 0
	err := RetryWithBackoff(context.Background(), 3, func() error {
		attempts++
		if attempts < 2 {
			return fmt.Errorf("temporary error")
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected no error after retry, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	err := RetryWithBackoff(context.Background(), 3, func() error {
		attempts++
		return errors.New("persistent error")
	})
	if err == nil {
		t.Errorf("expected error after max retries")
	}
	if attempts != 4 {
		t.Errorf("expected 4 attempts (1 initial + 3 retries), got %d", attempts)
	}
}

func TestRetryWithBackoff_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RetryWithBackoff(ctx, 3, func() error {
		return errors.New("some error")
	})
	if err == nil {
		t.Errorf("expected context error")
	}
}
