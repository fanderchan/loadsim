package stress

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

const dutyScale = 10000

type CPUScope string

const (
	ScopeWorkers CPUScope = "workers"
	ScopeHost    CPUScope = "host"
)

type CPUIdleMode string

const (
	IdleModePark CPUIdleMode = "park"
	IdleModeTrim CPUIdleMode = "trim"
)

type CPUConfig struct {
	Mode            Mode
	Scope           CPUScope
	IdleMode        CPUIdleMode
	Percent         float64
	MinPercent      float64
	MaxPercent      float64
	Period          time.Duration
	Cores           int
	Cycle           time.Duration
	ControlInterval time.Duration
	SampleDuration  time.Duration
	DeadbandPercent float64
	MaxStepPercent  float64
}

type CPUStatus struct {
	Mode             Mode
	Scope            CPUScope
	IdleMode         CPUIdleMode
	ActiveWorkers    int
	MaxWorkers       int
	RequestedPercent float64
	AppliedPercent   float64
	LastHostPercent  float64
}

type CPUStressor struct {
	config CPUConfig

	lock             sync.RWMutex
	stopOnce         sync.Once
	stopCh           chan struct{}
	workerWG         sync.WaitGroup
	controllerWG     sync.WaitGroup
	workers          []*cpuWorker
	running          bool
	startedAt        time.Time
	requestedPercent float64
	appliedPercent   float64
	lastHostPercent  float64
}

type cpuWorker struct {
	duty atomic.Uint32
	stop chan struct{}
}

func NewCPUStressor(config CPUConfig) (*CPUStressor, error) {
	if config.Cores <= 0 {
		config.Cores = runtime.NumCPU()
	}
	if maxCores := runtime.NumCPU(); config.Cores > maxCores {
		config.Cores = maxCores
	}
	if config.Cycle <= 0 {
		config.Cycle = 100 * time.Millisecond
	}
	if config.ControlInterval <= 0 {
		config.ControlInterval = 250 * time.Millisecond
	}
	if config.SampleDuration <= 0 {
		config.SampleDuration = 200 * time.Millisecond
	}
	if config.DeadbandPercent <= 0 {
		config.DeadbandPercent = 1.0
	}
	if config.MaxStepPercent <= 0 {
		config.MaxStepPercent = 10.0
	}
	if config.Scope == "" {
		config.Scope = ScopeWorkers
	}
	if config.IdleMode == "" {
		config.IdleMode = IdleModePark
	}
	if config.Cores <= 0 {
		return nil, fmt.Errorf("worker core count must be greater than zero")
	}
	if config.Scope != ScopeWorkers && config.Scope != ScopeHost {
		return nil, fmt.Errorf("CPU scope must be workers or host")
	}
	if config.IdleMode != IdleModePark && config.IdleMode != IdleModeTrim {
		return nil, fmt.Errorf("CPU idle mode must be park or trim")
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

	maxHostPercent := maxReachableHostPercent(config.Cores, runtime.NumCPU())
	if config.Scope == ScopeHost {
		switch config.Mode {
		case ModeFixed:
			if config.Percent > maxHostPercent {
				return nil, fmt.Errorf(
					"unreachable host CPU target %.1f%%: %d workers can provide at most %.2f%% on a %d-core host",
					config.Percent,
					config.Cores,
					maxHostPercent,
					runtime.NumCPU(),
				)
			}
		case ModeWave:
			if config.MaxPercent > maxHostPercent {
				return nil, fmt.Errorf(
					"unreachable host CPU wave max %.1f%%: %d workers can provide at most %.2f%% on a %d-core host",
					config.MaxPercent,
					config.Cores,
					maxHostPercent,
					runtime.NumCPU(),
				)
			}
		}
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
	s.workers = make([]*cpuWorker, 0, s.config.Cores)
	if s.config.IdleMode == IdleModePark {
		s.ensureWorkersLocked(s.config.Cores)
	}

	switch s.config.Mode {
	case ModeFixed:
		s.requestedPercent = s.config.Percent
		if s.config.Scope == ScopeWorkers {
			s.applyTargetLocked(s.config.Percent)
		}
	case ModeWave:
		s.requestedPercent = s.config.MinPercent
		if s.config.Scope == ScopeWorkers {
			s.applyTargetLocked(s.config.MinPercent)
		}
	}

	if s.config.Scope == ScopeHost {
		initialRequested := s.currentRequestedPercent()
		if hostUsage, err := sampleHostCPUPercent(s.config.SampleDuration); err == nil {
			s.requestedPercent = initialRequested
			s.lastHostPercent = hostUsage
			initialDriveHostPercent := clampFloat(initialRequested-hostUsage, 0, maxReachableHostPercent(s.config.Cores, runtime.NumCPU()))
			initialAppliedPercent := hostPercentToWorkerPercent(initialDriveHostPercent, s.config.Cores, runtime.NumCPU())
			s.applyTargetLocked(initialAppliedPercent)
		}
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
		s.requestedPercent = 0
		s.appliedPercent = 0
		s.lastHostPercent = 0
		s.workers = nil
	})
	return nil
}

func (s *CPUStressor) Status() CPUStatus {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return CPUStatus{
		Mode:             s.config.Mode,
		Scope:            s.config.Scope,
		IdleMode:         s.config.IdleMode,
		ActiveWorkers:    activeWorkerCount(s.appliedPercent, s.config.Cores),
		MaxWorkers:       s.config.Cores,
		RequestedPercent: s.requestedPercent,
		AppliedPercent:   s.appliedPercent,
		LastHostPercent:  s.lastHostPercent,
	}
}

func (s *CPUStressor) controlLoop() {
	defer s.controllerWG.Done()

	tick := s.config.ControlInterval
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.controlTick()
		}
	}
}

func (s *CPUStressor) controlTick() {
	requested := s.currentRequestedPercent()

	if s.config.Scope == ScopeWorkers {
		s.lock.Lock()
		s.requestedPercent = requested
		s.applyTargetLocked(requested)
		s.lock.Unlock()
		return
	}

	hostUsage, err := sampleHostCPUPercent(s.config.SampleDuration)
	if err != nil {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.requestedPercent = requested
	s.lastHostPercent = hostUsage

	nextAppliedPercent := nextHostAdaptiveAppliedPercent(
		requested,
		hostUsage,
		s.appliedPercent,
		s.config.Cores,
		runtime.NumCPU(),
		s.config.DeadbandPercent,
		s.config.MaxStepPercent,
	)
	s.applyTargetLocked(nextAppliedPercent)
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
	s.appliedPercent = percent

	activeUnits := float64(s.config.Cores) * percent / 100
	fullWorkers := int(math.Floor(activeUnits))
	partial := activeUnits - float64(fullWorkers)
	requiredWorkers := fullWorkers
	if partial > 0 {
		requiredWorkers++
	}

	switch s.config.IdleMode {
	case IdleModeTrim:
		s.ensureWorkersLocked(requiredWorkers)
		s.trimWorkersLocked(requiredWorkers)
	default:
		s.ensureWorkersLocked(s.config.Cores)
	}

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

func (s *CPUStressor) ensureWorkersLocked(count int) {
	if count < 0 {
		count = 0
	}
	if count > s.config.Cores {
		count = s.config.Cores
	}
	for len(s.workers) < count {
		worker := &cpuWorker{stop: make(chan struct{})}
		s.workers = append(s.workers, worker)
		s.workerWG.Add(1)
		go func(w *cpuWorker) {
			defer s.workerWG.Done()
			runCPUWorker(s.stopCh, w, s.config.Cycle)
		}(worker)
	}
}

func (s *CPUStressor) trimWorkersLocked(count int) {
	if count < 0 {
		count = 0
	}
	for len(s.workers) > count {
		last := s.workers[len(s.workers)-1]
		close(last.stop)
		s.workers = s.workers[:len(s.workers)-1]
	}
}

func runCPUWorker(stop <-chan struct{}, worker *cpuWorker, cycle time.Duration) {
	for {
		select {
		case <-stop:
			return
		case <-worker.stop:
			return
		default:
		}

		duty := worker.duty.Load()
		switch {
		case duty == 0:
			if !sleepOrStop(stop, worker.stop, cycle) {
				return
			}
		case duty >= dutyScale:
			if !busyUntil(stop, worker.stop, time.Now().Add(cycle)) {
				return
			}
		default:
			busyFor := time.Duration(int64(cycle) * int64(duty) / dutyScale)
			if busyFor > 0 && !busyUntil(stop, worker.stop, time.Now().Add(busyFor)) {
				return
			}
			if rest := cycle - busyFor; rest > 0 && !sleepOrStop(stop, worker.stop, rest) {
				return
			}
		}
	}
}

func busyUntil(stop <-chan struct{}, workerStop <-chan struct{}, deadline time.Time) bool {
	var sink float64
	for time.Now().Before(deadline) {
		select {
		case <-stop:
			return false
		case <-workerStop:
			return false
		default:
		}

		sink += math.Sqrt(12345.6789)
		sink += math.Sin(sink)
	}
	runtime.KeepAlive(sink)
	return true
}

func sleepOrStop(stop <-chan struct{}, workerStop <-chan struct{}, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-stop:
		return false
	case <-workerStop:
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

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func maxReachableHostPercent(workers, hostCPUs int) float64 {
	if workers <= 0 || hostCPUs <= 0 {
		return 0
	}
	return clampFloat(float64(workers)*100/float64(hostCPUs), 0, 100)
}

func workerPercentToHostPercent(workerPercent float64, workers, hostCPUs int) float64 {
	return clampFloat(workerPercent*float64(workers)/float64(hostCPUs), 0, 100)
}

func hostPercentToWorkerPercent(hostPercent float64, workers, hostCPUs int) float64 {
	if workers <= 0 || hostCPUs <= 0 {
		return 0
	}
	return clampFloat(hostPercent*float64(hostCPUs)/float64(workers), 0, 100)
}

func sampleHostCPUPercent(sampleDuration time.Duration) (float64, error) {
	percentages, err := cpu.Percent(sampleDuration, false)
	if err != nil {
		return 0, err
	}
	if len(percentages) == 0 {
		return 0, fmt.Errorf("failed to sample host CPU percent")
	}
	return percentages[0], nil
}

func activeWorkerCount(appliedPercent float64, cores int) int {
	if cores <= 0 {
		return 0
	}

	activeUnits := float64(cores) * clampFloat(appliedPercent, 0, 100) / 100
	fullWorkers := int(math.Floor(activeUnits))
	partial := activeUnits - float64(fullWorkers)
	if partial > 0 {
		fullWorkers++
	}
	if fullWorkers > cores {
		return cores
	}
	return fullWorkers
}

func (s *CPUStressor) currentRequestedPercent() float64 {
	if s.config.Mode == ModeWave {
		return s.wavePercent(time.Since(s.startedAt))
	}
	return s.config.Percent
}

func nextHostAdaptiveAppliedPercent(
	requestedHostPercent float64,
	observedHostPercent float64,
	currentAppliedPercent float64,
	workers int,
	hostCPUs int,
	deadbandPercent float64,
	maxStepPercent float64,
) float64 {
	requestedHostPercent = clampFloat(requestedHostPercent, 0, 100)
	observedHostPercent = clampFloat(observedHostPercent, 0, 100)
	currentAppliedPercent = clampFloat(currentAppliedPercent, 0, 100)
	maxStepPercent = clampFloat(maxStepPercent, 0, 100)

	currentDriveHostPercent := workerPercentToHostPercent(currentAppliedPercent, workers, hostCPUs)
	if math.Abs(observedHostPercent-requestedHostPercent) <= deadbandPercent {
		return currentAppliedPercent
	}

	nextDriveHostPercent := requestedHostPercent - maxFloat(observedHostPercent-currentDriveHostPercent, 0)
	nextDriveHostPercent = clampFloat(nextDriveHostPercent, 0, maxReachableHostPercent(workers, hostCPUs))

	desiredAppliedPercent := hostPercentToWorkerPercent(nextDriveHostPercent, workers, hostCPUs)
	delta := desiredAppliedPercent - currentAppliedPercent
	if delta > maxStepPercent {
		delta = maxStepPercent
	} else if delta < -maxStepPercent {
		delta = -maxStepPercent
	}

	return clampFloat(currentAppliedPercent+delta, 0, 100)
}
