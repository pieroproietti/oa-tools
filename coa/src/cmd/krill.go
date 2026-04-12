package cmd

import (
	"coa/src/internal/krill"

	"github.com/spf13/cobra"
)

var krillCmd = &cobra.Command{
	Use:   "krill",
	Short: "Start the system installation (The Hatching)",
	Long: `Krill is the interactive system installer for coa. 
It utilizes a TUI (Text User Interface) to gather user preferences and orchestrates the physical installation of the live environment onto a target disk.

⚠️ WARNING: This operation is destructive. It will zap the partition table of the target disk, create an EFI/ROOT layout, and format the partitions.`,
	Example: `  # Launch the interactive installer
  sudo coa krill`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), true)
		krill.HandleKrill()
	},
}

func init() {
	rootCmd.AddCommand(krillCmd)
}
