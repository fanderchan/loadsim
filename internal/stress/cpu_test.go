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
				Scope:   ScopeWorkers,
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
				Scope:      ScopeWorkers,
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
				Scope:      ScopeWorkers,
				MinPercent: 20,
				MaxPercent: 80,
			},
			wantErr: true,
		},
		{
			name: "host fixed unreachable target",
			cfg: CPUConfig{
				Mode:     ModeFixed,
				Scope:    ScopeHost,
				IdleMode: IdleModePark,
				Percent:  100,
				Cores:    1,
			},
			wantErr: true,
		},
		{
			name: "invalid idle mode",
			cfg: CPUConfig{
				Mode:     ModeFixed,
				Scope:    ScopeWorkers,
				IdleMode: CPUIdleMode("drop"),
				Percent:  10,
				Cores:    1,
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
		Scope:      ScopeWorkers,
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

func TestCPUPercentConversionHelpers(t *testing.T) {
	if got := maxReachableHostPercent(4, 160); got != 2.5 {
		t.Fatalf("maxReachableHostPercent got %.2f want 2.50", got)
	}

	if got := workerPercentToHostPercent(50, 4, 160); got != 1.25 {
		t.Fatalf("workerPercentToHostPercent got %.2f want 1.25", got)
	}

	if got := hostPercentToWorkerPercent(50, 160, 160); got != 50 {
		t.Fatalf("hostPercentToWorkerPercent got %.2f want 50.00", got)
	}
}

func TestNextHostAdaptiveAppliedPercent(t *testing.T) {
	tests := []struct {
		name     string
		target   float64
		observed float64
		current  float64
		deadband float64
		maxStep  float64
		want     float64
	}{
		{
			name:     "backs off to estimated remaining drive when host already exceeds target",
			target:   50,
			observed: 70,
			current:  40,
			deadband: 1,
			maxStep:  100,
			want:     20,
		},
		{
			name:     "drops to zero when baseline host load already exceeds target",
			target:   50,
			observed: 70,
			current:  10,
			deadband: 1,
			maxStep:  100,
			want:     0,
		},
		{
			name:     "holds within deadband",
			target:   50,
			observed: 50.5,
			current:  40,
			deadband: 1,
			maxStep:  100,
			want:     40,
		},
		{
			name:     "limits adjustment step",
			target:   50,
			observed: 10,
			current:  0,
			deadband: 1,
			maxStep:  10,
			want:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextHostAdaptiveAppliedPercent(tt.target, tt.observed, tt.current, 160, 160, tt.deadband, tt.maxStep)
			if got != tt.want {
				t.Fatalf("got %.2f want %.2f", got, tt.want)
			}
		})
	}
}

func TestTrimIdleModeShrinksWorkerPool(t *testing.T) {
	stressor, err := NewCPUStressor(CPUConfig{
		Mode:     ModeFixed,
		Scope:    ScopeWorkers,
		IdleMode: IdleModeTrim,
		Percent:  0,
		Cores:    4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer stressor.Stop()

	stressor.lock.Lock()
	stressor.applyTargetLocked(50)
	activeAfterGrow := len(stressor.workers)
	stressor.applyTargetLocked(0)
	activeAfterTrim := len(stressor.workers)
	stressor.lock.Unlock()

	if activeAfterGrow != 2 {
		t.Fatalf("active workers after grow = %d want 2", activeAfterGrow)
	}
	if activeAfterTrim != 0 {
		t.Fatalf("active workers after trim = %d want 0", activeAfterTrim)
	}
}

func TestCPUStatusCountsOnlyActiveWorkers(t *testing.T) {
	stressor, err := NewCPUStressor(CPUConfig{
		Mode:     ModeFixed,
		Scope:    ScopeWorkers,
		IdleMode: IdleModePark,
		Percent:  0,
		Cores:    4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer stressor.Stop()

	stressor.lock.Lock()
	stressor.applyTargetLocked(50)
	stressor.lock.Unlock()

	status := stressor.Status()
	if status.ActiveWorkers != 2 {
		t.Fatalf("status active workers = %d want 2", status.ActiveWorkers)
	}
	if status.MaxWorkers != 4 {
		t.Fatalf("status max workers = %d want 4", status.MaxWorkers)
	}
}
