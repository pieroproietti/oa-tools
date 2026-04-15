package cmd

import (
	"coa/src/internal/distro"
	"coa/src/internal/engine"

	"github.com/spf13/cobra"
)

var sysinstallCmd = &cobra.Command{
	Use:   "sysinstall",
	Short: "Install the live system to a physical disk",
	Long: `The 'sysinstall' command is the interactive system installer.
It gathers user preferences via TUI and orchestrates the physical installation 
on the target disk.
WARNING: This operation is destructive.`,
	Example: `  # Launch the system installer
  sudo coa sysinstall`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Verifica i privilegi di root (necessari per partizionare e montare)
		CheckSudoRequirements(cmd.Name(), true)

		// 2. Rileva la distribuzione host (necessaria per i comandi shell di krill)
		myDistro := distro.NewDistro()

		// 3. Passa la distro rilevata a HandleKrill
		engine.HandleKrill(myDistro)
	},
}

func init() {
	rootCmd.AddCommand(sysinstallCmd)
}
