package cmd

import (
	"fmt"
	"time"

	"github.com/fanderchan/loadsim/internal/stress"
	"github.com/fanderchan/loadsim/internal/system"

	"github.com/spf13/cobra"
)

var (
	ramMode           string
	ramSizeMB         int
	ramMinSizeMB      int
	ramMaxSizeMB      int
	ramWavePeriodSec  int
	ramBlockMB        int
	ramControlMS      int
	ramRateLimitMB    int
	ramRunTimeSec     int
	ramStatusEverySec int
)

var ramCmd = &cobra.Command{
	Use:   "ram",
	Short: "Occupy RAM with fixed or wave patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := stress.ParseMode(ramMode)
		if err != nil {
			return err
		}

		cfg := stress.RAMConfig{
			Mode:              mode,
			SizeMB:            ramSizeMB,
			MinSizeMB:         ramMinSizeMB,
			MaxSizeMB:         ramMaxSizeMB,
			Period:            time.Duration(ramWavePeriodSec) * time.Second,
			BlockMB:           ramBlockMB,
			ControlInterval:   time.Duration(ramControlMS) * time.Millisecond,
			RateLimitMBPerSec: ramRateLimitMB,
		}

		stressor, err := stress.NewRAMStressor(cfg)
		if err != nil {
			return err
		}

		if err := stressor.Start(); err != nil {
			return err
		}

		reason := watchLoop(
			time.Duration(ramRunTimeSec)*time.Second,
			time.Duration(ramStatusEverySec)*time.Second,
			func() { printRAMStatus(stressor) },
		)

		if err := stressor.Stop(); err != nil {
			return err
		}

		fmt.Printf("stopped: %s\n", reason)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ramCmd)

	ramCmd.Flags().StringVar(&ramMode, "mode", "fixed", "fixed or wave")
	ramCmd.Flags().IntVar(&ramSizeMB, "size", 1024, "fixed RAM target in MB")
	ramCmd.Flags().IntVar(&ramMinSizeMB, "min-size", 256, "wave mode minimum RAM in MB")
	ramCmd.Flags().IntVar(&ramMaxSizeMB, "max-size", 1024, "wave mode maximum RAM in MB")
	ramCmd.Flags().IntVar(&ramWavePeriodSec, "period", 60, "wave mode period in seconds")
	ramCmd.Flags().IntVar(&ramBlockMB, "block-size", 16, "RAM allocation block size in MB")
	ramCmd.Flags().IntVar(&ramControlMS, "control-ms", 250, "RAM control interval in milliseconds")
	ramCmd.Flags().IntVar(&ramRateLimitMB, "rate-limit", 0, "RAM change rate limit in MB per second, 0 means unlimited")
	ramCmd.Flags().IntVar(&ramRunTimeSec, "time", 0, "run time in seconds, 0 means no limit")
	ramCmd.Flags().IntVar(&ramStatusEverySec, "status-interval", 2, "status print interval in seconds")
}

func printRAMStatus(stressor *stress.RAMStressor) {
	status := stressor.Status()
	stats, err := system.Snapshot(150 * time.Millisecond)
	if err != nil {
		fmt.Printf(
			"[%s] ram mode=%s target=%dMB current=%dMB block=%dMB rate_limit=%dMB/s\n",
			time.Now().Format("15:04:05"),
			status.Mode,
			status.TargetMB,
			status.CurrentMB,
			status.BlockMB,
			status.RateLimitMB,
		)
		return
	}

	fmt.Printf(
		"[%s] ram mode=%s target=%dMB current=%dMB block=%dMB rate_limit=%dMB/s host_cpu=%.1f%% host_mem=%.1f%%\n",
		time.Now().Format("15:04:05"),
		status.Mode,
		status.TargetMB,
		status.CurrentMB,
		status.BlockMB,
		status.RateLimitMB,
		stats.CPUPercent,
		stats.MemoryPercent,
	)
}
