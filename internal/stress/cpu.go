package stress

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const dutyScale = 10000

type CPUConfig struct {
	Mode       Mode
	Percent    float64
	MinPercent float64
	MaxPercent float64
	Period     time.Duration
	Cores      int
	Cycle      time.Duration
}

type CPUStatus struct {
	Mode          Mode
	Cores         int
	TargetPercent float64
}

type CPUStressor struct {
	config CPUConfig

	lock          sync.RWMutex
	stopOnce      sync.Once
	stopCh        chan struct{}
	workerWG      sync.WaitGroup
	controllerWG  sync.WaitGroup
	workers       []*cpuWorker
	running       bool
	startedAt     time.Time
	targetPercent float64
}

type cpuWorker struct {
	duty atomic.Uint32
}

func NewCPUStressor(config CPUConfig) (*CPUStressor, error) {
	if config.Cores <= 0 {
		config.Cores = runtime.NumCPU()
	}
	if config.Cycle <= 0 {
		config.Cycle = 100 * time.Millisecond
	}
	if config.Cores <= 0 {
		return nil, fmt.Errorf("worker core count must be greater than zero")
	}

	switch config.Mode {
	case ModeFixed:
		if config.Percent < 0 || config.Percent > 100 {
			return nil, fmt.Errorf("CPU percent must be between 0 and 100")
		}
	case ModeWave:
		if config.MinPercent < 0 || config.MaxPercent > 100 {
			return nil, fmt.Errorf("CPU wave percent must stay between 0 and 100")
		}
		if config.MinPercent > config.MaxPercent {
			return nil, fmt.Errorf("CPU min percent must be less than or equal to max percent")
		}
		if config.Period <= 0 {
			return nil, fmt.Errorf("CPU wave period must be greater than zero")
		}
	default:
		return nil, fmt.Errorf("unsupported CPU mode %q", config.Mode)
	}

	return &CPUStressor{
		config: config,
		stopCh: make(chan struct{}),
	}, nil
}

func (s *CPUStressor) Start() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.running {
		return fmt.Errorf("CPU stressor is already running")
	}

	s.running = true
	s.startedAt = time.Now()
	s.workers = make([]*cpuWorker, s.config.Cores)
	for i := 0; i < s.config.Cores; i++ {
		worker := &cpuWorker{}
		s.workers[i] = worker
		s.workerWG.Add(1)
		go func(w *cpuWorker) {
			defer s.workerWG.Done()
			runCPUWorker(s.stopCh, w, s.config.Cycle)
		}(worker)
	}

	switch s.config.Mode {
	case ModeFixed:
		s.applyTargetLocked(s.config.Percent)
	case ModeWave:
		s.applyTargetLocked(s.config.MinPercent)
	}

	s.controllerWG.Add(1)
	go s.controlLoop()
	return nil
}

func (s *CPUStressor) Stop() error {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		s.controllerWG.Wait()
		s.workerWG.Wait()

		s.lock.Lock()
		defer s.lock.Unlock()

		s.running = false
		s.targetPercent = 0
		s.workers = nil
	})
	return nil
}

func (s *CPUStressor) Status() CPUStatus {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return CPUStatus{
		Mode:          s.config.Mode,
		Cores:         s.config.Cores,
		TargetPercent: s.targetPercent,
	}
}

func (s *CPUStressor) controlLoop() {
	defer s.controllerWG.Done()

	tick := 250 * time.Millisecond
	if s.config.Mode == ModeWave {
		candidate := s.config.Period / 40
		if candidate >= 50*time.Millisecond && candidate < tick {
			tick = candidate
		}
	}

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if s.config.Mode != ModeWave {
				continue
			}
			target := s.wavePercent(time.Since(s.startedAt))
			s.lock.Lock()
			s.applyTargetLocked(target)
			s.lock.Unlock()
		}
	}
}

func (s *CPUStressor) wavePercent(elapsed time.Duration) float64 {
	phase := math.Mod(elapsed.Seconds(), s.config.Period.Seconds()) / s.config.Period.Seconds()
	span := s.config.MaxPercent - s.config.MinPercent
	if phase < 0.5 {
		return s.config.MinPercent + span*phase*2
	}
	return s.config.MaxPercent - span*(phase-0.5)*2
}

func (s *CPUStressor) applyTargetLocked(percent float64) {
	percent = clampFloat(percent, 0, 100)
	s.targetPercent = percent

	activeUnits := float64(s.config.Cores) * percent / 100
	fullWorkers := int(math.Floor(activeUnits))
	partial := activeUnits - float64(fullWorkers)

	for idx, worker := range s.workers {
		switch {
		case idx < fullWorkers:
			worker.duty.Store(dutyScale)
		case idx == fullWorkers && partial > 0 && fullWorkers < len(s.workers):
			worker.duty.Store(uint32(math.Round(partial * dutyScale)))
		default:
			worker.duty.Store(0)
		}
	}
}

func runCPUWorker(stop <-chan struct{}, worker *cpuWorker, cycle time.Duration) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		duty := worker.duty.Load()
		switch {
		case duty == 0:
			if !sleepOrStop(stop, cycle) {
				return
			}
		case duty >= dutyScale:
			if !busyUntil(stop, time.Now().Add(cycle)) {
				return
			}
		default:
			busyFor := time.Duration(int64(cycle) * int64(duty) / dutyScale)
			if busyFor > 0 && !busyUntil(stop, time.Now().Add(busyFor)) {
				return
			}
			if rest := cycle - busyFor; rest > 0 && !sleepOrStop(stop, rest) {
				return
			}
		}
	}
}

func busyUntil(stop <-chan struct{}, deadline time.Time) bool {
	var sink float64
	for time.Now().Before(deadline) {
		select {
		case <-stop:
			return false
		default:
		}

		sink += math.Sqrt(12345.6789)
		sink += math.Sin(sink)
	}
	runtime.KeepAlive(sink)
	return true
}

func sleepOrStop(stop <-chan struct{}, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-stop:
		return false
	case <-timer.C:
		return true
	}
}

func clampFloat(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
