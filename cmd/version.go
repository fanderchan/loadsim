package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.2.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("LoadSim %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
