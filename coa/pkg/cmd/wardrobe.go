package cmd

import (
	"github.com/spf13/cobra"
)

var (
	distroFlag string
	jsonFlag   bool
	noAccFlag  bool
	noFirmFlag bool
)

var wardrobeCmd = &cobra.Command{
	Use:   "wardrobe",
	Short: "Gestisce e applica vestiti (costumi e accessori) al sistema",
}

func init() {
	// Aggiungi solo i comandi
	wardrobeCmd.AddCommand(getCmd, listCmd, showCmd, wearCmd)

	// NON definire i flag qui se li definisci nei file dedicati!

	rootCmd.AddCommand(wardrobeCmd)
}
