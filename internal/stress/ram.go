package stress

import (
	"fmt"
	"math"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type RAMConfig struct {
	Mode              Mode
	SizeMB            int
	MinSizeMB         int
	MaxSizeMB         int
	Period            time.Duration
	BlockMB           int
	ControlInterval   time.Duration
	RateLimitMBPerSec int
}

type RAMStatus struct {
	Mode        Mode
	TargetMB    int
	AppliedMB   int
	CurrentMB   int
	BlockMB     int
	RateLimitMB int
}

type ramBlock struct {
	sizeMB int
	data   []byte
}

type RAMStressor struct {
	config RAMConfig

	lock      sync.RWMutex
	stopOnce  sync.Once
	stopCh    chan struct{}
	wg        sync.WaitGroup
	running   bool
	startedAt time.Time
	targetMB  int
	currentMB int
	blocks    []ramBlock
}

func NewRAMStressor(config RAMConfig) (*RAMStressor, error) {
	if config.BlockMB <= 0 {
		config.BlockMB = 16
	}
	if config.ControlInterval <= 0 {
		config.ControlInterval = 250 * time.Millisecond
	}

	switch config.Mode {
	case ModeFixed:
		if config.SizeMB <= 0 {
			return nil, fmt.Errorf("RAM size must be greater than zero")
		}
	case ModeWave:
		if config.MinSizeMB < 0 || config.MaxSizeMB <= 0 {
			return nil, fmt.Errorf("RAM wave bounds must be non-negative and max must be greater than zero")
		}
		if config.MinSizeMB > config.MaxSizeMB {
			return nil, fmt.Errorf("RAM min size must be less than or equal to max size")
		}
		if config.Period <= 0 {
			return nil, fmt.Errorf("RAM wave period must be greater than zero")
		}
	default:
		return nil, fmt.Errorf("unsupported RAM mode %q", config.Mode)
	}

	return &RAMStressor{
		config: config,
		stopCh: make(chan struct{}),
	}, nil
}

func (s *RAMStressor) Start() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.running {
		return fmt.Errorf("RAM stressor is already running")
	}

	s.running = true
	s.startedAt = time.Now()

	s.wg.Add(1)
	go s.controlLoop()
	return nil
}

func (s *RAMStressor) Stop() error {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		s.wg.Wait()

		s.lock.Lock()
		defer s.lock.Unlock()

		for idx := range s.blocks {
			s.blocks[idx].data = nil
		}
		s.blocks = nil
		s.currentMB = 0
		s.targetMB = 0
		s.running = false

		runtime.GC()
		debug.FreeOSMemory()
	})
	return nil
}

func (s *RAMStressor) Status() RAMStatus {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return RAMStatus{
		Mode:        s.config.Mode,
		TargetMB:    s.targetMB,
		AppliedMB:   s.currentMB,
		CurrentMB:   s.currentMB,
		BlockMB:     s.config.BlockMB,
		RateLimitMB: s.config.RateLimitMBPerSec,
	}
}

func (s *RAMStressor) controlLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.ControlInterval)
	defer ticker.Stop()

	s.applyDesiredTarget()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.applyDesiredTarget()
		}
	}
}

func (s *RAMStressor) desiredTarget() int {
	if s.config.Mode == ModeWave {
		return s.config.MinSizeMB
	}
	return s.config.SizeMB
}

func (s *RAMStressor) applyDesiredTarget() {
	target := s.desiredTarget()
	if s.config.Mode == ModeWave {
		target = s.waveTarget(time.Since(s.startedAt))
	}
	target = s.limitTargetChange(target)
	s.resizeTo(target)
}

func (s *RAMStressor) waveTarget(elapsed time.Duration) int {
	phase := math.Mod(elapsed.Seconds(), s.config.Period.Seconds()) / s.config.Period.Seconds()
	span := float64(s.config.MaxSizeMB - s.config.MinSizeMB)
	if phase < 0.5 {
		return s.config.MinSizeMB + int(math.Round(span*phase*2))
	}
	return s.config.MaxSizeMB - int(math.Round(span*(phase-0.5)*2))
}

func (s *RAMStressor) resizeTo(targetMB int) {
	if targetMB < 0 {
		targetMB = 0
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.targetMB = targetMB

	for s.currentMB < targetMB {
		chunkMB := s.config.BlockMB
		remaining := targetMB - s.currentMB
		if remaining < chunkMB {
			chunkMB = remaining
		}

		block := ramBlock{
			sizeMB: chunkMB,
			data:   allocateRAMBlock(chunkMB),
		}
		s.blocks = append(s.blocks, block)
		s.currentMB += chunkMB
	}

	for s.currentMB > targetMB && len(s.blocks) > 0 {
		lastIdx := len(s.blocks) - 1
		last := s.blocks[lastIdx]
		excess := s.currentMB - targetMB

		switch {
		case excess >= last.sizeMB:
			s.blocks[lastIdx].data = nil
			s.blocks = s.blocks[:lastIdx]
			s.currentMB -= last.sizeMB
		default:
			newSize := last.sizeMB - excess
			s.blocks[lastIdx].data = allocateRAMBlock(newSize)
			s.blocks[lastIdx].sizeMB = newSize
			s.currentMB = targetMB
		}
	}
}

func (s *RAMStressor) limitTargetChange(targetMB int) int {
	if s.config.RateLimitMBPerSec <= 0 {
		return targetMB
	}

	s.lock.RLock()
	currentMB := s.currentMB
	s.lock.RUnlock()

	maxDelta := int(math.Ceil(float64(s.config.RateLimitMBPerSec) * s.config.ControlInterval.Seconds()))
	if maxDelta < 1 {
		maxDelta = 1
	}

	switch {
	case targetMB > currentMB+maxDelta:
		return currentMB + maxDelta
	case targetMB < currentMB-maxDelta:
		return currentMB - maxDelta
	default:
		return targetMB
	}
}

func allocateRAMBlock(sizeMB int) []byte {
	if sizeMB <= 0 {
		return nil
	}

	block := make([]byte, sizeMB*1024*1024)
	for offset := 0; offset < len(block); offset += 4096 {
		block[offset] = 1
	}
	if len(block) > 0 {
		block[len(block)-1] = 1
	}
	return block
}
