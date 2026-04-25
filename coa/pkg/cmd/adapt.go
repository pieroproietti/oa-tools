package cmd

import (
	"os/exec"

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
		handleAdapt()
	},
}

func init() {
	rootCmd.AddCommand(adaptCmd)
}

// =====================================================================
// LOGICA DI ADATTAMENTO (Ex-Engine)
// =====================================================================

// handleAdapt adatta la risoluzione del monitor per le Virtual Machine
func handleAdapt() {
	// Nota: usiamo LogCoala e LogSuccess che abbiamo già definito nel pacchetto cmd (es. in remaster.go)
	LogCoala("Adapting monitor resolution...")

	virtualOutputs := []string{"Virtual-0", "Virtual-1", "Virtual-2", "Virtual-3"}

	for _, output := range virtualOutputs {
		cmd := exec.Command("xrandr", "--output", output, "--auto")
		_ = cmd.Run() // Ignoriamo gli errori in modo silenzioso, perché non tutti gli output esisteranno
	}

	LogSuccess("Resolution adapted.")
}
