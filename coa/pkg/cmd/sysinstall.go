package cmd

import (
	"coa/pkg/pilot"
	"coa/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

var sysinstallCmd = &cobra.Command{
	Use:   "sysinstall",
	Short: "Lancia l'installatore di sistema (GUI o TUI)",
	Run: func(cmd *cobra.Command, args []string) {
		// Controllo permessi (coa deve girare come root)
		CheckSudoRequirements(cmd.Name(), true)

		// 1. Caricamento del profilo tramite il Pilot
		profile, err := pilot.DetectAndLoad()
		if err != nil {
			utils.LogError("Impossibile caricare il profilo: %v", err)
			os.Exit(1)
		}

		// 2. Scelta dell'installatore (per ora forziamo Calamares)
		// Qui passiamo il 'profile' che abbiamo appena caricato!
		RunCalamaresInstaller(profile)
	},
}

func init() {
	rootCmd.AddCommand(sysinstallCmd)
}
