package cmd

import (
	"coa/pkg/tailor"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Elenca costumi e accessori presenti nel wardrobe",
	Args:  cobra.NoArgs, // Rimosso l'argomento [REPO_NAME]
	Run: func(cmd *cobra.Command, args []string) {
		// Passiamo solo distroFlag. Il sarto sa già di guardare in ~/.wardrobe
		if err := tailor.List(distroFlag); err != nil {
			fmt.Println("❌ Errore durante il List:", err)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&distroFlag, "distro", "d", "", "Filtra per distribuzione")
}
