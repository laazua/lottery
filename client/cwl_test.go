package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCWLClient_FetchDraws_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"total": 1,
			"data": [{
				"issue": "24180",
				"drawTime": "2026-07-04 21:20:00",
				"frontWinningNum": "05,12,18,23,31",
				"backWinningNum": "07,11",
				"saleAmount": "310000000",
				"poolAmount": "920000000"
			}]
		}`))
	}))
	defer server.Close()

	c := NewCWLClient(WithBaseURL(server.URL))
	draws, err := c.FetchDraws(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(draws) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(draws))
	}
	if draws[0].Issue != "24180" {
		t.Errorf("expected issue 24180, got %s", draws[0].Issue)
	}
	if draws[0].FrontNumbers != [5]int{5, 12, 18, 23, 31} {
		t.Errorf("unexpected front numbers: %v", draws[0].FrontNumbers)
	}
	if draws[0].BackNumbers != [2]int{7, 11} {
		t.Errorf("unexpected back numbers: %v", draws[0].BackNumbers)
	}
}

func TestCWLClient_FetchDraws_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"total":0,"data":[]}`))
	}))
	defer server.Close()

	c := NewCWLClient(WithBaseURL(server.URL))
	draws, err := c.FetchDraws(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(draws) != 0 {
		t.Errorf("expected empty draws, got %d", len(draws))
	}
}

func TestCWLClient_FetchDraws_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewCWLClient(WithBaseURL(server.URL))
	draws, err := c.FetchDraws(context.Background())
	if err != nil {
		t.Fatalf("API失败自动回退到模拟数据应无错误: %v", err)
	}
	if len(draws) == 0 {
		t.Error("回退的模拟数据不应为空")
	}
}

func TestCWLClient_FetchDraws_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := NewCWLClient(WithBaseURL(server.URL))
	draws, err := c.FetchDraws(context.Background())
	if err != nil {
		t.Fatalf("API限流自动回退到模拟数据应无错误: %v", err)
	}
	if len(draws) == 0 {
		t.Error("回退的模拟数据不应为空")
	}
}

func TestParseDrawResponse_Valid(t *testing.T) {
	data := []byte(`{
		"total": 1,
		"data": [{
			"issue": "24180",
			"drawTime": "2026-07-04 21:20:00",
			"frontWinningNum": "05,12,18,23,31",
			"backWinningNum": "07,11",
			"saleAmount": "310000000",
			"poolAmount": "920000000"
		}]
	}`)

	draws, err := parseDrawResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(draws) != 1 {
		t.Fatalf("expected 1 draw, got %d", len(draws))
	}
}

func TestParseDrawResponse_EmptyData(t *testing.T) {
	data := []byte(`{"total":0,"data":[]}`)
	draws, err := parseDrawResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(draws) != 0 {
		t.Errorf("expected 0 draws, got %d", len(draws))
	}
}

func TestParseDrawResponse_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json}`)
	_, err := parseDrawResponse(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseDrawResponse_EmptyBody(t *testing.T) {
	_, err := parseDrawResponse([]byte{})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestCWLClient_FetchDrawByPeriod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"total": 1,
			"data": [{
				"issue": "24180",
				"drawTime": "2026-07-04 21:20:00",
				"frontWinningNum": "05,12,18,23,31",
				"backWinningNum": "07,11",
				"saleAmount": "310000000",
				"poolAmount": "920000000"
			}]
		}`))
	}))
	defer server.Close()

	c := NewCWLClient(WithBaseURL(server.URL))
	draw, err := c.FetchDrawByPeriod(context.Background(), "24180")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draw.Issue != "24180" {
		t.Errorf("expected issue 24180, got %s", draw.Issue)
	}
}

func TestOptions_WithPageSize(t *testing.T) {
	o := newOptions()
	WithPageSize(50)(o)
	if o.pageSize != 50 {
		t.Errorf("expected pageSize 50, got %d", o.pageSize)
	}
}

func TestOptions_WithTimeout(t *testing.T) {
	o := newOptions()
	WithTimeout(30 * time.Second)(o)
	if o.timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", o.timeout)
	}
}
