package cmd

import (
	"coa/src/internal/engine"

	"github.com/spf13/cobra"
)

var adaptCmd = &cobra.Command{
	Use:   "adapt",
	Short: "Adapt the video resolution to the Virtual Machine window",
	Long: `The 'adapt' command is a post-boot utility specifically designed for live environments running inside Virtual Machines (such as VirtualBox, VMware, or QEMU/KVM). 

When executed, it forces the guest operating system to dynamically resize its display resolution to perfectly match the current dimensions of the host's VM window, improving the user experience during testing.`,
	Example: `  # Automatically resize the live system display to fit the VM window
  coa adapt`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), false)
		engine.HandleAdapt()
	},
}

func init() {
	rootCmd.AddCommand(adaptCmd)
}
