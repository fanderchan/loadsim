package stress

import (
	"testing"
	"time"
)

func TestNewRAMStressorValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     RAMConfig
		wantErr bool
	}{
		{
			name: "fixed ok",
			cfg: RAMConfig{
				Mode:   ModeFixed,
				SizeMB: 64,
			},
		},
		{
			name: "fixed invalid size",
			cfg: RAMConfig{
				Mode:   ModeFixed,
				SizeMB: 0,
			},
			wantErr: true,
		},
		{
			name: "wave invalid bounds",
			cfg: RAMConfig{
				Mode:      ModeWave,
				MinSizeMB: 128,
				MaxSizeMB: 64,
				Period:    60 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "wave invalid period",
			cfg: RAMConfig{
				Mode:      ModeWave,
				MinSizeMB: 64,
				MaxSizeMB: 128,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRAMStressor(tt.cfg)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRAMWaveTarget(t *testing.T) {
	stressor, err := NewRAMStressor(RAMConfig{
		Mode:      ModeWave,
		MinSizeMB: 64,
		MaxSizeMB: 256,
		Period:    60 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []struct {
		elapsed time.Duration
		want    int
	}{
		{elapsed: 0, want: 64},
		{elapsed: 15 * time.Second, want: 160},
		{elapsed: 30 * time.Second, want: 256},
		{elapsed: 45 * time.Second, want: 160},
	}

	for _, tc := range cases {
		got := stressor.waveTarget(tc.elapsed)
		if got != tc.want {
			t.Fatalf("elapsed=%v got=%d want=%d", tc.elapsed, got, tc.want)
		}
	}
}

func TestRAMResizeToExactTarget(t *testing.T) {
	stressor, err := NewRAMStressor(RAMConfig{
		Mode:    ModeFixed,
		SizeMB:  1,
		BlockMB: 16,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stressor.resizeTo(20)
	if stressor.currentMB != 20 {
		t.Fatalf("currentMB=%d want=20 after grow", stressor.currentMB)
	}

	stressor.resizeTo(5)
	if stressor.currentMB != 5 {
		t.Fatalf("currentMB=%d want=5 after shrink", stressor.currentMB)
	}

	stressor.resizeTo(0)
	if stressor.currentMB != 0 {
		t.Fatalf("currentMB=%d want=0 after clear", stressor.currentMB)
	}
}

func TestRAMRateLimitCapsTargetChange(t *testing.T) {
	stressor, err := NewRAMStressor(RAMConfig{
		Mode:              ModeWave,
		MinSizeMB:         0,
		MaxSizeMB:         256,
		Period:            60 * time.Second,
		ControlInterval:   250 * time.Millisecond,
		RateLimitMBPerSec: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stressor.currentMB = 10
	if got := stressor.limitTargetChange(100); got != 15 {
		t.Fatalf("limitTargetChange grow got %d want 15", got)
	}

	stressor.currentMB = 100
	if got := stressor.limitTargetChange(0); got != 95 {
		t.Fatalf("limitTargetChange shrink got %d want 95", got)
	}
}

func TestRAMApplyDesiredTargetRespectsStartupRateLimit(t *testing.T) {
	stressor, err := NewRAMStressor(RAMConfig{
		Mode:              ModeFixed,
		SizeMB:            128,
		ControlInterval:   250 * time.Millisecond,
		RateLimitMBPerSec: 16,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stressor.applyDesiredTarget()
	if stressor.currentMB != 4 {
		t.Fatalf("currentMB=%d want=4 after startup-limited growth", stressor.currentMB)
	}
	if stressor.targetMB != 4 {
		t.Fatalf("targetMB=%d want=4 after startup-limited growth", stressor.targetMB)
	}
}
