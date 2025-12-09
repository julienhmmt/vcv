package config

import (
	"os"
	"testing"
)

func TestLoadExpirationThresholds_Defaults(t *testing.T) {
	// Clear environment variables
	if err := os.Unsetenv("VCV_EXPIRE_CRITICAL"); err != nil {
		t.Fatalf("failed to unset VCV_EXPIRE_CRITICAL: %v", err)
	}
	if err := os.Unsetenv("VCV_EXPIRE_WARNING"); err != nil {
		t.Fatalf("failed to unset VCV_EXPIRE_WARNING: %v", err)
	}

	thresholds := loadExpirationThresholds()

	if thresholds.Critical != 7 {
		t.Errorf("expected default critical 7, got %d", thresholds.Critical)
	}
	if thresholds.Warning != 30 {
		t.Errorf("expected default warning 30, got %d", thresholds.Warning)
	}
}

func TestLoadExpirationThresholds_CustomValues(t *testing.T) {
	if err := os.Setenv("VCV_EXPIRE_CRITICAL", "14"); err != nil {
		t.Fatalf("failed to set VCV_EXPIRE_CRITICAL: %v", err)
	}
	if err := os.Setenv("VCV_EXPIRE_WARNING", "60"); err != nil {
		t.Fatalf("failed to set VCV_EXPIRE_WARNING: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("VCV_EXPIRE_CRITICAL"); err != nil {
			t.Fatalf("failed to unset VCV_EXPIRE_CRITICAL: %v", err)
		}
		if err := os.Unsetenv("VCV_EXPIRE_WARNING"); err != nil {
			t.Fatalf("failed to unset VCV_EXPIRE_WARNING: %v", err)
		}
	}()

	thresholds := loadExpirationThresholds()

	if thresholds.Critical != 14 {
		t.Errorf("expected critical 14, got %d", thresholds.Critical)
	}
	if thresholds.Warning != 60 {
		t.Errorf("expected warning 60, got %d", thresholds.Warning)
	}
}

func TestLoadExpirationThresholds_InvalidValues(t *testing.T) {
	if err := os.Setenv("VCV_EXPIRE_CRITICAL", "invalid"); err != nil {
		t.Fatalf("failed to set VCV_EXPIRE_CRITICAL: %v", err)
	}
	if err := os.Setenv("VCV_EXPIRE_WARNING", "not_a_number"); err != nil {
		t.Fatalf("failed to set VCV_EXPIRE_WARNING: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("VCV_EXPIRE_CRITICAL"); err != nil {
			t.Fatalf("failed to unset VCV_EXPIRE_CRITICAL: %v", err)
		}
		if err := os.Unsetenv("VCV_EXPIRE_WARNING"); err != nil {
			t.Fatalf("failed to unset VCV_EXPIRE_WARNING: %v", err)
		}
	}()

	thresholds := loadExpirationThresholds()

	// Should fall back to defaults on invalid input
	if thresholds.Critical != 7 {
		t.Errorf("expected default critical 7 on invalid input, got %d", thresholds.Critical)
	}
	if thresholds.Warning != 30 {
		t.Errorf("expected default warning 30 on invalid input, got %d", thresholds.Warning)
	}
}

func TestLoadExpirationThresholds_PartialCustom(t *testing.T) {
	if err := os.Setenv("VCV_EXPIRE_CRITICAL", "21"); err != nil {
		t.Fatalf("failed to set VCV_EXPIRE_CRITICAL: %v", err)
	}
	if err := os.Unsetenv("VCV_EXPIRE_WARNING"); err != nil {
		t.Fatalf("failed to unset VCV_EXPIRE_WARNING: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("VCV_EXPIRE_CRITICAL"); err != nil {
			t.Fatalf("failed to unset VCV_EXPIRE_CRITICAL: %v", err)
		}
	}()

	thresholds := loadExpirationThresholds()

	if thresholds.Critical != 21 {
		t.Errorf("expected critical 21, got %d", thresholds.Critical)
	}
	if thresholds.Warning != 30 {
		t.Errorf("expected default warning 30, got %d", thresholds.Warning)
	}
}
