package cmd

import (
	"fmt"
	"time"

	"github.com/fanderchan/loadsim/internal/stress"
	"github.com/fanderchan/loadsim/internal/system"

	"github.com/spf13/cobra"
)

var (
	comboCPUPercent     float64
	comboCPUMinPercent  float64
	comboCPUMaxPercent  float64
	comboCPUPeriodSec   int
	comboCPUCores       int
	comboCPUMode        string
	comboCPUScope       string
	comboCPUIdleMode    string
	comboCPUControlMS   int
	comboCPUSampleMS    int
	comboCPUDeadband    float64
	comboCPUMaxStep     float64
	comboRAMMode        string
	comboRAMSizeMB      int
	comboRAMMinSizeMB   int
	comboRAMMaxSizeMB   int
	comboRAMPeriodSec   int
	comboRAMBlockMB     int
	comboRAMControlMS   int
	comboRAMRateLimitMB int
	comboRunTimeSec     int
	comboStatusEverySec int
)

var comboCmd = &cobra.Command{
	Use:   "combo",
	Short: "Occupy CPU and RAM at the same time",
	RunE: func(cmd *cobra.Command, args []string) error {
		cpuMode, err := stress.ParseMode(comboCPUMode)
		if err != nil {
			return err
		}
		ramMode, err := stress.ParseMode(comboRAMMode)
		if err != nil {
			return err
		}

		cpuCfg := stress.CPUConfig{
			Mode:            cpuMode,
			Scope:           stress.CPUScope(comboCPUScope),
			IdleMode:        stress.CPUIdleMode(comboCPUIdleMode),
			Percent:         comboCPUPercent,
			MinPercent:      comboCPUMinPercent,
			MaxPercent:      comboCPUMaxPercent,
			Period:          time.Duration(comboCPUPeriodSec) * time.Second,
			Cores:           comboCPUCores,
			ControlInterval: time.Duration(comboCPUControlMS) * time.Millisecond,
			SampleDuration:  time.Duration(comboCPUSampleMS) * time.Millisecond,
			DeadbandPercent: comboCPUDeadband,
			MaxStepPercent:  comboCPUMaxStep,
		}
		ramCfg := stress.RAMConfig{
			Mode:              ramMode,
			SizeMB:            comboRAMSizeMB,
			MinSizeMB:         comboRAMMinSizeMB,
			MaxSizeMB:         comboRAMMaxSizeMB,
			Period:            time.Duration(comboRAMPeriodSec) * time.Second,
			BlockMB:           comboRAMBlockMB,
			ControlInterval:   time.Duration(comboRAMControlMS) * time.Millisecond,
			RateLimitMBPerSec: comboRAMRateLimitMB,
		}

		if !cpuConfigActive(cpuCfg) || !ramConfigActive(ramCfg) {
			return fmt.Errorf("combo requires both CPU and RAM targets to be greater than zero")
		}

		cpuStressor, err := stress.NewCPUStressor(cpuCfg)
		if err != nil {
			return err
		}
		ramStressor, err := stress.NewRAMStressor(ramCfg)
		if err != nil {
			return err
		}

		if err := cpuStressor.Start(); err != nil {
			return err
		}
		if err := ramStressor.Start(); err != nil {
			_ = cpuStressor.Stop()
			return err
		}

		reason := watchLoop(
			time.Duration(comboRunTimeSec)*time.Second,
			time.Duration(comboStatusEverySec)*time.Second,
			func() { printComboStatus(cpuStressor, ramStressor) },
		)

		if err := joinErrors(ramStressor.Stop(), cpuStressor.Stop()); err != nil {
			return err
		}

		fmt.Printf("stopped: %s\n", reason)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(comboCmd)

	comboCmd.Flags().StringVar(&comboCPUMode, "cpu-mode", "fixed", "CPU mode: fixed or wave")
	comboCmd.Flags().StringVar(&comboCPUScope, "cpu-scope", "workers", "CPU target scope: workers or host")
	comboCmd.Flags().StringVar(&comboCPUIdleMode, "cpu-idle-mode", "park", "CPU idle worker behavior: park or trim")
	comboCmd.Flags().Float64Var(&comboCPUPercent, "cpu-percent", 50, "fixed CPU target percent")
	comboCmd.Flags().Float64Var(&comboCPUMinPercent, "cpu-min", 20, "wave CPU minimum percent")
	comboCmd.Flags().Float64Var(&comboCPUMaxPercent, "cpu-max", 80, "wave CPU maximum percent")
	comboCmd.Flags().IntVar(&comboCPUPeriodSec, "cpu-period", 60, "wave CPU period in seconds")
	comboCmd.Flags().IntVar(&comboCPUCores, "cpu-cores", 0, "worker core count, 0 uses all host cores")
	comboCmd.Flags().IntVar(&comboCPUControlMS, "cpu-control-ms", 250, "CPU controller adjustment interval in milliseconds")
	comboCmd.Flags().IntVar(&comboCPUSampleMS, "cpu-sample-ms", 200, "CPU host sample duration in milliseconds")
	comboCmd.Flags().Float64Var(&comboCPUDeadband, "cpu-deadband", 1.0, "CPU host deadband percent before adjusting")
	comboCmd.Flags().Float64Var(&comboCPUMaxStep, "cpu-max-step", 10.0, "maximum CPU worker drive change per control step in percent")

	comboCmd.Flags().StringVar(&comboRAMMode, "ram-mode", "fixed", "RAM mode: fixed or wave")
	comboCmd.Flags().IntVar(&comboRAMSizeMB, "ram-size", 1024, "fixed RAM target in MB")
	comboCmd.Flags().IntVar(&comboRAMMinSizeMB, "ram-min-size", 256, "wave RAM minimum in MB")
	comboCmd.Flags().IntVar(&comboRAMMaxSizeMB, "ram-max-size", 1024, "wave RAM maximum in MB")
	comboCmd.Flags().IntVar(&comboRAMPeriodSec, "ram-period", 60, "wave RAM period in seconds")
	comboCmd.Flags().IntVar(&comboRAMBlockMB, "ram-block-size", 16, "RAM allocation block size in MB")
	comboCmd.Flags().IntVar(&comboRAMControlMS, "ram-control-ms", 250, "RAM control interval in milliseconds")
	comboCmd.Flags().IntVar(&comboRAMRateLimitMB, "ram-rate-limit", 0, "RAM change rate limit in MB per second, 0 means unlimited")

	comboCmd.Flags().IntVar(&comboRunTimeSec, "time", 0, "run time in seconds, 0 means no limit")
	comboCmd.Flags().IntVar(&comboStatusEverySec, "status-interval", 2, "status print interval in seconds")
}

func cpuConfigActive(cfg stress.CPUConfig) bool {
	switch cfg.Mode {
	case stress.ModeFixed:
		return cfg.Percent > 0
	case stress.ModeWave:
		return cfg.MaxPercent > 0
	default:
		return false
	}
}

func ramConfigActive(cfg stress.RAMConfig) bool {
	switch cfg.Mode {
	case stress.ModeFixed:
		return cfg.SizeMB > 0
	case stress.ModeWave:
		return cfg.MaxSizeMB > 0
	default:
		return false
	}
}

func printComboStatus(cpuStressor *stress.CPUStressor, ramStressor *stress.RAMStressor) {
	cpuStatus := cpuStressor.Status()
	ramStatus := ramStressor.Status()
	stats, err := system.Snapshot(150 * time.Millisecond)
	if err != nil {
		fmt.Printf(
			"[%s] combo cpu_scope=%s cpu_idle=%s cpu_target=%.1f%% cpu_drive=%.1f%% workers=%d/%d ram=%dMB/%dMB\n",
			time.Now().Format("15:04:05"),
			cpuStatus.Scope,
			cpuStatus.IdleMode,
			cpuStatus.RequestedPercent,
			cpuStatus.AppliedPercent,
			cpuStatus.ActiveWorkers,
			cpuStatus.MaxWorkers,
			ramStatus.CurrentMB,
			ramStatus.TargetMB,
		)
		return
	}

	fmt.Printf(
		"[%s] combo cpu_scope=%s cpu_idle=%s cpu_target=%.1f%% cpu_drive=%.1f%% workers=%d/%d ram_target=%dMB ram_current=%dMB host_cpu=%.1f%% host_mem=%.1f%%\n",
		time.Now().Format("15:04:05"),
		cpuStatus.Scope,
		cpuStatus.IdleMode,
		cpuStatus.RequestedPercent,
		cpuStatus.AppliedPercent,
		cpuStatus.ActiveWorkers,
		cpuStatus.MaxWorkers,
		ramStatus.TargetMB,
		ramStatus.CurrentMB,
		stats.CPUPercent,
		stats.MemoryPercent,
	)
}
