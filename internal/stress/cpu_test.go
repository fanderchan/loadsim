package stress

import (
	"testing"
	"time"
)

func TestNewCPUStressorValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     CPUConfig
		wantErr bool
	}{
		{
			name: "fixed ok",
			cfg: CPUConfig{
				Mode:    ModeFixed,
				Percent: 50,
				Cores:   2,
			},
		},
		{
			name: "fixed invalid percent",
			cfg: CPUConfig{
				Mode:    ModeFixed,
				Percent: 120,
			},
			wantErr: true,
		},
		{
			name: "wave invalid bounds",
			cfg: CPUConfig{
				Mode:       ModeWave,
				MinPercent: 80,
				MaxPercent: 20,
				Period:     60 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "wave invalid period",
			cfg: CPUConfig{
				Mode:       ModeWave,
				MinPercent: 20,
				MaxPercent: 80,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCPUStressor(tt.cfg)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCPUWavePercent(t *testing.T) {
	stressor, err := NewCPUStressor(CPUConfig{
		Mode:       ModeWave,
		MinPercent: 20,
		MaxPercent: 80,
		Period:     60 * time.Second,
		Cores:      1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []struct {
		elapsed time.Duration
		want    float64
	}{
		{elapsed: 0, want: 20},
		{elapsed: 15 * time.Second, want: 50},
		{elapsed: 30 * time.Second, want: 80},
		{elapsed: 45 * time.Second, want: 50},
	}

	for _, tc := range cases {
		got := stressor.wavePercent(tc.elapsed)
		if got != tc.want {
			t.Fatalf("elapsed=%v got=%.1f want=%.1f", tc.elapsed, got, tc.want)
		}
	}
}
