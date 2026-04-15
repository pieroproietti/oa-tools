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

var remasterCmd = &cobra.Command{
	Use:   "remaster",
	Short: "Start a system remastering flight (ISO production)",
	Long: `The 'remaster' command orchestrates the creation of a bootable live ISO. 
It uses OverlayFS for a zero-copy approach and produces artifacts ready for distribution.`,
	Example: `  # Start a standard ISO remastering
  sudo coa remaster --mode standard`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), true)
		myDistro := distro.NewDistro()
		engine.HandleRemaster(produceMode, producePath, myDistro) // Il gestore interno può restare Produce o essere rinominato in HandleRemaster
	},
}

func init() {
	remasterCmd.Flags().StringVar(&produceMode, "mode", "standard", "standard, clone, or crypted")
	remasterCmd.Flags().StringVar(&producePath, "path", "/home/eggs", "working directory")

	rootCmd.AddCommand(remasterCmd)
}
