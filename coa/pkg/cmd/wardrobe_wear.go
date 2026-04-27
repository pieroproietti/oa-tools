package cmd

import (
	"coa/pkg/tailor"
	"coa/pkg/utils" // Usiamo utils per i log colorati
	"os"

	"github.com/spf13/cobra"
)

var wearCmd = &cobra.Command{
	Use:   "wear [COSTUME]",
	Short: "Indossa un costume o accessorio dal wardrobe (richiede sudo)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 🚨 CONTROLLO PRIVILEGI: Se non sei root, ci fermiamo subito.
		if os.Geteuid() != 0 {
			utils.LogError("Questo comando deve essere eseguito con privilegi di root.")
			utils.LogCoala("Prova a digitare: sudo coa wardrobe wear %s", args[0])
			os.Exit(1)
		}

		costume := args[0]

		// Se arriviamo qui, siamo root.
		if err := tailor.Wear(costume, noAccFlag, noFirmFlag); err != nil {
			utils.LogError("Errore durante il Wear: %v", err)
			os.Exit(1)
		}
	},
}
