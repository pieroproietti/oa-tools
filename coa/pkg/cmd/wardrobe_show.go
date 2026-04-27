package cmd

import (
	"coa/pkg/tailor"
	"fmt"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show [COSTUME]",
	Short: "Mostra i dettagli di un costume o accessorio",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		costume := args[0]
		// jsonFlag è accessibile perché definita in wardrobe.go nello stesso package
		if err := tailor.Show(costume, jsonFlag); err != nil {
			fmt.Println("❌ Errore durante lo Show:", err)
		}
	},
}
