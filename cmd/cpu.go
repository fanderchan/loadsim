package cmd

import (
	"fmt"
	"time"

	"github.com/fanderchan/loadsim/internal/stress"
	"github.com/fanderchan/loadsim/internal/system"

	"github.com/spf13/cobra"
)

var (
	cpuMode           string
	cpuScope          string
	cpuIdleMode       string
	cpuPercent        float64
	cpuMinPercent     float64
	cpuMaxPercent     float64
	cpuWavePeriodSec  int
	cpuCores          int
	cpuControlMS      int
	cpuSampleMS       int
	cpuDeadband       float64
	cpuMaxStep        float64
	cpuRunTimeSec     int
	cpuStatusEverySec int
)

var cpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: "Occupy CPU with fixed or wave patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := stress.ParseMode(cpuMode)
		if err != nil {
			return err
		}

		cfg := stress.CPUConfig{
			Mode:            mode,
			Scope:           stress.CPUScope(cpuScope),
			IdleMode:        stress.CPUIdleMode(cpuIdleMode),
			Percent:         cpuPercent,
			MinPercent:      cpuMinPercent,
			MaxPercent:      cpuMaxPercent,
			Period:          time.Duration(cpuWavePeriodSec) * time.Second,
			Cores:           cpuCores,
			ControlInterval: time.Duration(cpuControlMS) * time.Millisecond,
			SampleDuration:  time.Duration(cpuSampleMS) * time.Millisecond,
			DeadbandPercent: cpuDeadband,
			MaxStepPercent:  cpuMaxStep,
		}

		stressor, err := stress.NewCPUStressor(cfg)
		if err != nil {
			return err
		}

		if err := stressor.Start(); err != nil {
			return err
		}

		reason := watchLoop(
			time.Duration(cpuRunTimeSec)*time.Second,
			time.Duration(cpuStatusEverySec)*time.Second,
			func() { printCPUStatus(stressor) },
		)

		if err := stressor.Stop(); err != nil {
			return err
		}

		fmt.Printf("stopped: %s\n", reason)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cpuCmd)

	cpuCmd.Flags().StringVar(&cpuMode, "mode", "fixed", "fixed or wave")
	cpuCmd.Flags().StringVar(&cpuScope, "scope", "workers", "CPU target scope: workers or host")
	cpuCmd.Flags().StringVar(&cpuIdleMode, "idle-mode", "park", "idle worker behavior: park or trim")
	cpuCmd.Flags().Float64Var(&cpuPercent, "percent", 50, "fixed CPU target percent")
	cpuCmd.Flags().Float64Var(&cpuMinPercent, "min", 20, "wave mode minimum CPU percent")
	cpuCmd.Flags().Float64Var(&cpuMaxPercent, "max", 80, "wave mode maximum CPU percent")
	cpuCmd.Flags().IntVar(&cpuWavePeriodSec, "period", 60, "wave mode period in seconds")
	cpuCmd.Flags().IntVar(&cpuCores, "cores", 0, "worker core count, 0 uses all host cores")
	cpuCmd.Flags().IntVar(&cpuControlMS, "control-ms", 250, "controller adjustment interval in milliseconds")
	cpuCmd.Flags().IntVar(&cpuSampleMS, "sample-ms", 200, "host CPU sample duration in milliseconds")
	cpuCmd.Flags().Float64Var(&cpuDeadband, "deadband", 1.0, "host CPU deadband percent before adjusting")
	cpuCmd.Flags().Float64Var(&cpuMaxStep, "max-step", 10.0, "maximum worker drive change per control step in percent")
	cpuCmd.Flags().IntVar(&cpuRunTimeSec, "time", 0, "run time in seconds, 0 means no limit")
	cpuCmd.Flags().IntVar(&cpuStatusEverySec, "status-interval", 2, "status print interval in seconds")
}

func printCPUStatus(stressor *stress.CPUStressor) {
	status := stressor.Status()
	stats, err := system.Snapshot(150 * time.Millisecond)
	if err != nil {
		fmt.Printf(
			"[%s] cpu mode=%s scope=%s idle=%s target=%.1f%% drive=%.1f%% workers=%d/%d\n",
			time.Now().Format("15:04:05"),
			status.Mode,
			status.Scope,
			status.IdleMode,
			status.RequestedPercent,
			status.AppliedPercent,
			status.ActiveWorkers,
			status.MaxWorkers,
		)
		return
	}

	fmt.Printf(
		"[%s] cpu mode=%s scope=%s idle=%s target=%.1f%% drive=%.1f%% workers=%d/%d host_cpu=%.1f%% host_mem=%.1f%%\n",
		time.Now().Format("15:04:05"),
		status.Mode,
		status.Scope,
		status.IdleMode,
		status.RequestedPercent,
		status.AppliedPercent,
		status.ActiveWorkers,
		status.MaxWorkers,
		stats.CPUPercent,
		stats.MemoryPercent,
	)
}
