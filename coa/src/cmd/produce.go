package cmd

import (
	"coa/src/internal/distro"
	"coa/src/internal/engine"

	"github.com/spf13/cobra"
)

var (
	produceMode string
	producePath string
)

var produceCmd = &cobra.Command{
	Use:   "produce",
	Short: "Start a system remastering production flight",
	Long: `The 'produce' command is the core of the remastering process. 
It orchestrates the creation of a bootable live ISO using OverlayFS for a zero-copy approach.

Supported modes:
  - standard: Creates a fresh live system (purging host user identities).
  - clone: Preserves the host's /home directory and user identities.
  - crypted: Encapsulates the root filesystem inside a LUKS2 container.`,
	Example: `  # Start a standard ISO production in the default /home/eggs nest
  sudo coa produce

  # Start a clone production in a custom path
  sudo coa produce --mode clone --path /mnt/storage/workspace

  # Start an encrypted live production
  sudo coa produce --mode crypted`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), true)
		myDistro := distro.NewDistro()
		engine.HandleProduce(produceMode, producePath, myDistro)
	},
}

func init() {
	produceCmd.Flags().StringVar(&produceMode, "mode", "standard", "standard, clone, or crypted")
	produceCmd.Flags().StringVar(&producePath, "path", "/home/eggs", "working directory")

	rootCmd.AddCommand(produceCmd)
}
