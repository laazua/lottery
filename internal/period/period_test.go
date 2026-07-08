package period

import (
	"testing"
)

func TestValidatePeriod_Valid(t *testing.T) {
	if err := ValidatePeriod("24180"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidatePeriod_TooShort(t *testing.T) {
	if err := ValidatePeriod("1234"); err == nil {
		t.Errorf("expected error for short period")
	}
}

func TestValidatePeriod_TooLong(t *testing.T) {
	if err := ValidatePeriod("123456"); err == nil {
		t.Errorf("expected error for long period")
	}
}

func TestValidatePeriod_NonNumeric(t *testing.T) {
	if err := ValidatePeriod("24xyz"); err == nil {
		t.Errorf("expected error for non-numeric period")
	}
}

func TestExtractYear(t *testing.T) {
	year, err := ExtractYear("24180")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if year != 2024 {
		t.Errorf("expected year 2024, got %d", year)
	}
}

func TestExtractYear_Minimal(t *testing.T) {
	year, err := ExtractYear("00001")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if year != 2000 {
		t.Errorf("expected year 2000, got %d", year)
	}
}

func TestExtractSequence(t *testing.T) {
	seq, err := ExtractSequence("24001")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if seq != 1 {
		t.Errorf("expected sequence 1, got %d", seq)
	}
}

func TestExtractSequence_Normal(t *testing.T) {
	seq, err := ExtractSequence("24180")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if seq != 180 {
		t.Errorf("expected sequence 180, got %d", seq)
	}
}

func TestExtractSequence_Zero(t *testing.T) {
	seq, err := ExtractSequence("24000")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if seq != 0 {
		t.Errorf("expected sequence 0, got %d", seq)
	}
}

func TestExtractYear_InvalidPeriod(t *testing.T) {
	_, err := ExtractYear("abc")
	if err == nil {
		t.Errorf("expected error for invalid period")
	}
}
