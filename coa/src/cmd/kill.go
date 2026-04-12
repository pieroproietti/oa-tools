package cmd

import (
	"coa/src/internal/engine"

	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Free the nest and unmount filesystems",
	Long: `Safely tears down the remastering environment. 
It uses MNT_DETACH to unmount the OverlayFS and virtual API filesystems (/dev, /proc, /sys) without affecting the running host, then removes the temporary workspace.`,
	Example: `  # Clean up the default workspace
  sudo coa kill`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), true)
		engine.HandleKill()
	},
}

func init() {
	rootCmd.AddCommand(killCmd)
}
