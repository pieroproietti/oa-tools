package cmd

import (
	"fmt"
	"os"

	"coa/pkg/bleach" // Assicurati che il path sia corretto

	"github.com/spf13/cobra"
)

var cleanVerbose bool

var toolsCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Pulisce log, cache apt/pacman e residui del sistema host",
	Long: `Dimagrisce il sistema rimuovendo file non necessari.
Ideale da lanciare prima di 'coa remaster' per ottenere una ISO più compatta.`,
	Example: `  sudo coa tools clean
  sudo coa tools clean --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		// Controllo root: la pulizia profonda richiede i privilegi massimi
		if os.Geteuid() != 0 {
			fmt.Println("\033[1;31m[ERRORE]\033[0m Il comando clean deve essere eseguito come root (sudo).")
			os.Exit(1)
		}

		fmt.Println("Inizio procedura di Bleach (pulizia profonda)...")

		b := bleach.New(cleanVerbose)
		if err := b.Clean(); err != nil {
			fmt.Printf("\033[1;31m[ERRORE]\033[0m Pulizia interrotta: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\033[1;32m[SUCCESS]\033[0m Sistema pulito! Ora la tua ISO sarà più snella.")
	},
}

func init() {
	// Aggiungiamo il comando 'clean' come figlio del comando 'tools'
	// Assicurati che toolsCmd sia dichiarato da qualche parte (es. in tools.go)
	toolsCmd.AddCommand(toolsCleanCmd)

	toolsCleanCmd.Flags().BoolVarP(&cleanVerbose, "verbose", "v", false, "Mostra l'output dettagliato dei comandi di pulizia")
}
