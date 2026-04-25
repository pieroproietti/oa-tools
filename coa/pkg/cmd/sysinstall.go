package cmd

import (
	"os"

	"coa/pkg/calamares"

	"github.com/spf13/cobra"
)

var sysinstallCmd = &cobra.Command{
	Use:   "sysinstall",
	Short: "Launch the graphical system installer (Calamares + OA)",
	Long: `The 'sysinstall' command prepares the environment and launches 
the Calamares graphical installer. Once the GUI finishes partitioning 
and unpacking, the OA engine will take over to finalize the bootloader.`,
	Example: `  # Launch the system installer
  sudo coa sysinstall`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckSudoRequirements(cmd.Name(), true)

		LogCoala("Preparazione ambiente e avvio di Calamares...")

		// Affidiamo tutto al configuratore nel nuovo pacchetto pkg/calamares
		err := calamares.SetupAndLaunch()
		if err != nil {
			LogError("L'installazione si è interrotta: %v", err)
			os.Exit(1)
		}

		LogSuccess("Installazione completata! Il sistema è pronto per il riavvio.")
	},
}

func init() {
	rootCmd.AddCommand(sysinstallCmd)
}
